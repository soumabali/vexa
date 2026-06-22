"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Trash2 } from "lucide-react"
import type { Tunnel } from "@/types/tunnel"
import { deleteTunnel } from "@/lib/api/tunnels"

interface TunnelDeleteDialogProps {
  tunnel: Tunnel
  onDeleted?: () => void
}

export function TunnelDeleteDialog({ tunnel, onDeleted }: TunnelDeleteDialogProps) {
  const [open, setOpen] = useState(false)
  const [confirmText, setConfirmText] = useState("")
  const [deleting, setDeleting] = useState(false)

  const expected = tunnel.interface_name
  const confirmed = confirmText === expected

  const handleDelete = async () => {
    if (!confirmed) return
    setDeleting(true)
    try {
      await deleteTunnel(tunnel.id)
      setOpen(false)
      onDeleted?.()
    } finally {
      setDeleting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="ghost" size="sm" className="text-destructive">
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Tunnel</DialogTitle>
          <DialogDescription>
            This will permanently delete the WireGuard tunnel{" "}
            <strong>{tunnel.interface_name}</strong>. This action cannot be
            undone.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>
              Type <strong>{expected}</strong> to confirm deletion
            </Label>
            <Input
              value={confirmText}
              onChange={(e) => setConfirmText(e.target.value)}
              placeholder={expected}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={!confirmed || deleting}
          >
            {deleting ? "Deleting…" : "Delete Tunnel"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
