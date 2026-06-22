# Windows Authenticode signing
param(
  [Parameter(Mandatory=$true)]
  [string]$File,
  [string]$CertificateThumbprint
)

if (-not $CertificateThumbprint) {
  $CertificateThumbprint = $env:WINDOWS_CERT_THUMB
}

Set-AuthenticodeSignature -FilePath $File -Certificate (Get-ChildItem Cert:\CurrentUser\My | Where-Object { $_.Thumbprint -eq $CertificateThumbprint }) -TimestampServer "http://timestamp.digicert.com"
Write-Host "Signed $File"
