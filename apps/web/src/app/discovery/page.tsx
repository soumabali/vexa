'use client'

import { useState, useEffect, useCallback } from 'react'
import { useAsyncData } from '@/hooks/useAsyncData'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import type { ScanJob, ScanResult, StartScanResponse, ScanResultsResponse, ScanHistoryResponse, ImportResultsResponse } from '@/types/discovery'
import { useToast } from '@/components/ui/use-toast'
import { api } from "@/lib/api"
import { MaterialIcon } from "@/components/ui/material-icon";

const defaultPorts = [
  { port: 22, service: 'SSH', checked: true },
  { port: 80, service: 'HTTP', checked: true },
  { port: 443, service: 'HTTPS', checked: true },
  { port: 3389, service: 'RDP', checked: true },
  { port: 5900, service: 'VNC', checked: true },
  { port: 5432, service: 'PostgreSQL', checked: false },
  { port: 3306, service: 'MySQL', checked: false },
  { port: 6379, service: 'Redis', checked: false },
  { port: 8080, service: 'HTTP-Alt', checked: false },
  { port: 8443, service: 'HTTPS-Alt', checked: false },
]

export default function DiscoveryPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [network, setNetwork] = useState('192.168.1.0/24')
  const [ports, setPorts] = useState(defaultPorts)
  const [options, setOptions] = useState({
    timeoutMs: 1000,
    rateLimitMs: 100,
    concurrency: 50,
    icmpFirst: true,
    osScan: true,
    resolveHostname: true,
  })
  const [isScanning, setIsScanning] = useState(false)
  const [currentScan, setCurrentScan] = useState<ScanJob | null>(null)
  const [scanResults, setScanResults] = useState<ScanResult[]>([])
  const [scanHistory, setScanHistory] = useState<ScanJob[]>([])
  const [selectedResults, setSelectedResults] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(false)

  const togglePort = (port: number) => {
    setPorts(ports.map(p => p.port === port ? { ...p, checked: !p.checked } : p))
  }

  const startScan = async () => {
    const selectedPorts = ports.filter(p => p.checked).map(p => p.port)
    if (selectedPorts.length === 0) {
      toast({ title: 'Error', description: 'Select at least one port to scan', variant: 'destructive' })
      return
    }

    setLoading(true)
    try {
      const response: StartScanResponse = await api<StartScanResponse>('/api/v1/discovery/scan', { method: 'POST', body: JSON.stringify({
        network,
        ports: selectedPorts,
        options: {
          timeout_ms: options.timeoutMs,
          rate_limit_ms: options.rateLimitMs,
          concurrency: options.concurrency,
          icmp_first: options.icmpFirst,
          os_scan: options.osScan,
          resolve_hostname: options.resolveHostname,
        },
      }) })

      const data = response
      setCurrentScan({
        id: data.id,
        network: data.network,
        ports: [],
        status: 'pending',
        progress: 0,
        results: [],
        total_hosts: 0,
        scanned_hosts: 0,
        created_at: data.created_at,
        options: { timeout_ms: 0, rate_limit_ms: 0, concurrency: 0, icmp_first: false, os_scan: false, resolve_hostname: false },
      })
      setIsScanning(true)
      toast({ title: 'Scan Started', description: `Scanning ${network}...` })

      // Start polling
      pollScanStatus(data.id)
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to start scan'
      toast({ title: 'Error', description: message, variant: 'destructive' })
    } finally {
      setLoading(false)
    }
  }

  const fetchScanResults = async (scanId: string) => {
    try {
      const response: ScanResultsResponse = await api<ScanResultsResponse>(`/discovery/scan/${scanId}/results`)
      setScanResults(response.results || [])
    } catch (error) {
      console.error('Failed to fetch results:', error)
    }
  }

  const { data: historyData, reload: refetchHistory } = useAsyncData<ScanJob[]>(async () => {
    const response: ScanHistoryResponse = await api<ScanHistoryResponse>('/api/v1/discovery/scan')
    return response.jobs || []
  })

  const pollScanStatus = useCallback(async (scanId: string) => {
    const interval = setInterval(async () => {
      try {
        const response: ScanJob = await api<ScanJob>(`/api/v1/discovery/scan/${scanId}/status`)
        const data = response
        setCurrentScan(data)

        if (data.status === 'completed' || data.status === 'cancelled' || data.status === 'failed') {
          clearInterval(interval)
          setIsScanning(false)
          if (data.status === 'completed') {
            fetchScanResults(scanId)
            toast({ title: 'Scan Complete', description: `Found ${data.found_hosts} hosts` })
          }
          refetchHistory()
        }
      } catch (error) {
        clearInterval(interval)
        setIsScanning(false)
      }
    }, 1000)

    return () => clearInterval(interval)
  }, [])

  const cancelScan = async (scanId: string) => {
    try {
      await api(`/api/v1/discovery/scan/${scanId}/cancel`, { method: 'POST' })
      toast({ title: 'Scan Cancelled' })
      setIsScanning(false)
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to cancel'
      toast({ title: 'Error', description: message, variant: 'destructive' })
    }
  }

  const importResults = async (scanId: string, importAll: boolean) => {
    try {
      const ips = importAll ? [] : Array.from(selectedResults)
      const response: ImportResultsResponse = await api<ImportResultsResponse>(`/api/v1/discovery/scan/${scanId}/import`, { method: 'POST', body: JSON.stringify({
        import_all: importAll,
        selected_ips: ips,
      }) })
      toast({ title: 'Import Successful', description: `Imported ${response.imported_count} hosts` })
      router.push('/hosts')
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to import'
      toast({ title: 'Error', description: message, variant: 'destructive' })
    }
  }

  const toggleResultSelection = (ip: string) => {
    const newSet = new Set(selectedResults)
    if (newSet.has(ip)) {
      newSet.delete(ip)
    } else {
      newSet.add(ip)
    }
    setSelectedResults(newSet)
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'up': return 'bg-green-500'
      case 'down': return 'bg-red-500'
      case 'filtered': return 'bg-yellow-500'
      default: return 'bg-gray-500'
    }
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'up': return <Badge variant="default" className="bg-green-600">Online</Badge>
      case 'down': return <Badge variant="destructive">Offline</Badge>
      case 'filtered': return <Badge variant="secondary">Filtered</Badge>
      default: return <Badge variant="outline">{status}</Badge>
    }
  }

  useEffect(() => {
    if (historyData) {
      Promise.resolve().then(() => {
        if (historyData) setScanHistory(historyData);
      });
    }
  }, [historyData]);

  return (
    <div className="container mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Network Discovery</h1>
        <div className="flex items-center gap-2 text-sm text-on-surface-variant">
          <MaterialIcon name="shield" className="h-4 w-4" />
          <span>Rate-limited, private networks only</span>
        </div>
      </div>

      {/* Scan Configuration */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <MaterialIcon name="search" className="h-5 w-5" />
            Scan Configuration
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Network Range (CIDR)</label>
              <Input
                value={network}
                onChange={(e) => setNetwork(e.target.value)}
                placeholder="192.168.1.0/24"
                disabled={isScanning}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Timeout (ms)</label>
              <Input
                type="number"
                value={options.timeoutMs}
                onChange={(e) => setOptions({ ...options, timeoutMs: parseInt(e.target.value) })}
                min={100}
                max={10000}
                disabled={isScanning}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Concurrency</label>
              <Input
                type="number"
                value={options.concurrency}
                onChange={(e) => setOptions({ ...options, concurrency: parseInt(e.target.value) })}
                min={1}
                max={200}
                disabled={isScanning}
              />
            </div>
          </div>

          {/* Port Selection */}
          <div className="space-y-2">
            <label className="text-sm font-medium">Ports to Scan</label>
            <div className="flex flex-wrap gap-3">
              {ports.map((port) => (
                <label key={port.port} className="flex items-center gap-2 cursor-pointer">
                  <Checkbox
                    checked={port.checked}
                    onCheckedChange={() => togglePort(port.port)}
                    disabled={isScanning}
                  />
                  <span className="text-sm">{port.port} ({port.service})</span>
                </label>
              ))}
            </div>
          </div>

          {/* Options */}
          <div className="flex flex-wrap gap-6">
            <label className="flex items-center gap-2 cursor-pointer">
              <Checkbox
                checked={options.icmpFirst}
                onCheckedChange={(checked) => setOptions({ ...options, icmpFirst: checked as boolean })}
                disabled={isScanning}
              />
              <span className="text-sm">ICMP ping first</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <Checkbox
                checked={options.osScan}
                onCheckedChange={(checked) => setOptions({ ...options, osScan: checked as boolean })}
                disabled={isScanning}
              />
              <span className="text-sm">OS fingerprinting</span>
            </label>
            <label className="flex items-center gap-2 cursor-pointer">
              <Checkbox
                checked={options.resolveHostname}
                onCheckedChange={(checked) => setOptions({ ...options, resolveHostname: checked as boolean })}
                disabled={isScanning}
              />
              <span className="text-sm">Resolve hostnames</span>
            </label>
          </div>

          {/* Action Buttons */}
          <div className="flex gap-3">
            <Button
              onClick={startScan}
              disabled={isScanning || loading}
              className="min-w-[120px]"
            >
              {isScanning ? (
                <MaterialIcon name="progress_activity" className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <MaterialIcon name="search" className="mr-2 h-4 w-4" />
              )}
              {isScanning ? 'Scanning...' : 'Start Scan'}
            </Button>
            {isScanning && currentScan && (
              <Button
                variant="outline"
                onClick={() => cancelScan(currentScan.id)}
                disabled={loading}
              >
                <MaterialIcon name="pause" className="mr-2 h-4 w-4" />
                Cancel
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Progress */}
      {isScanning && currentScan && (
        <Card>
          <CardHeader>
            <CardTitle>Scan Progress</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex justify-between text-sm">
              <span>{currentScan.network}</span>
              <span>{currentScan.scanned_hosts} / {currentScan.total_hosts} hosts</span>
            </div>
            <Progress value={currentScan.progress} className="h-2" />
            <div className="flex justify-between text-sm text-on-surface-variant">
              <span>Status: {currentScan.status}</span>
              <span>{currentScan.progress}% complete</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Results */}
      {scanResults.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span className="flex items-center gap-2">
                <MaterialIcon name="dns" className="h-5 w-5" />
                Scan Results ({scanResults.length} hosts found)
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => importResults(currentScan!.id, true)}
                >
                  <MaterialIcon name="add" className="mr-2 h-4 w-4" />
                  Import All
                </Button>
                {selectedResults.size > 0 && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => importResults(currentScan!.id, false)}
                  >
                    <MaterialIcon name="add" className="mr-2 h-4 w-4" />
                    Import Selected ({selectedResults.size})
                  </Button>
                )}
              </div>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
              <table className="w-full">
                <thead>
                  <tr className="border-b bg-surface-container-low/50">
                    <th className="p-3 text-left">
                      <Checkbox
                        checked={selectedResults.size === scanResults.length}
                        onCheckedChange={(checked) => {
                          if (checked) {
                            setSelectedResults(new Set(scanResults.map(r => r.ip_address)))
                          } else {
                            setSelectedResults(new Set())
                          }
                        }}
                      />
                    </th>
                    <th className="p-3 text-left">IP Address</th>
                    <th className="p-3 text-left">Hostname</th>
                    <th className="p-3 text-left">Status</th>
                    <th className="p-3 text-left">OS Guess</th>
                    <th className="p-3 text-left">Open Ports</th>
                    <th className="p-3 text-left">RTT</th>
                  </tr>
                </thead>
                <tbody>
                  {scanResults.map((result) => (
                    <tr key={result.id} className="border-b">
                      <td className="p-3">
                        <Checkbox
                          checked={selectedResults.has(result.ip_address)}
                          onCheckedChange={() => toggleResultSelection(result.ip_address)}
                        />
                      </td>
                      <td className="p-3 font-mono">{result.ip_address}</td>
                      <td className="p-3">{result.hostname || '-'}</td>
                      <td className="p-3">{getStatusBadge(result.status)}</td>
                      <td className="p-3">{result.os_guess || 'Unknown'}</td>
                      <td className="p-3">
                        <div className="flex flex-wrap gap-1">
                          {result.open_ports.map((port) => (
                            <Badge key={port.port} variant="outline" className="text-xs">
                              {port.port}/{port.protocol}
                              {port.service && ` (${port.service})`}
                              {port.tls && ' 🔒'}
                            </Badge>
                          ))}
                        </div>
                      </td>
                      <td className="p-3">{result.response_time_ms}ms</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Scan History */}
      {scanHistory.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Scan History</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {scanHistory.map((job) => (
                <div
                  key={job.id}
                  className="flex items-center justify-between p-3 rounded-md border hover:bg-surface-container-low/50 cursor-pointer"
                  onClick={() => {
                    if (job.status === 'completed') {
                      fetchScanResults(job.id)
                    }
                  }}
                >
                  <div className="flex items-center gap-3">
                    <MaterialIcon name="public" className="h-4 w-4 text-on-surface-variant" />
                    <div>
                      <p className="font-medium">{job.network}</p>
                      <p className="text-sm text-on-surface-variant">
                        {new Date(job.created_at).toLocaleString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right">
                      <p className="text-sm">{job.found_hosts} hosts found</p>
                      <p className="text-sm text-on-surface-variant">
                        {job.scanned_hosts}/{job.total_hosts} scanned
                      </p>
                    </div>
                    <Badge
                      variant={
                        job.status === 'completed' ? 'default' :
                        job.status === 'running' ? 'secondary' :
                        job.status === 'cancelled' ? 'outline' : 'destructive'
                      }
                    >
                      {job.status}
                    </Badge>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
