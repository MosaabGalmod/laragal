$certPath = "$env:APPDATA\caddy\pki\authorities\local\root.crt"
if (Test-Path $certPath) {
    Write-Output "Found root cert: $certPath"
    $cmd = "Import-Certificate -FilePath '$certPath' -CertStoreLocation 'Cert:\LocalMachine\Root'"
    Start-Process powershell -ArgumentList "-NoProfile -Command `"$cmd`"" -Verb RunAs -Wait
    Write-Output "CERT_INSTALL_COMPLETED"
} else {
    Write-Output "CERT_NOT_FOUND"
}
