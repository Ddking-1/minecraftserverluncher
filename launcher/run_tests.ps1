Set-Location 'D:/projektek/ai/minecraftserverluncher/launcher'
$exeDir = 'build-clean6/Release'
Get-ChildItem -Path $exeDir -Filter 'test_*.exe' | ForEach-Object {
    $name = $_.Name
    Write-Output ''
    Write-Output "=== $name ==="
    $output = & $_.FullName 2>&1
    Write-Output $output
    Write-Output "EXIT $name $LASTEXITCODE"
}
