# PowerShell Profile Setup Helper
# Este arquivo configura o PowerShell para permitir scripts locais

# Este é um script helper. Para usar:
# 1. Abra PowerShell como Admin
# 2. Execute: Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope CurrentUser
# 3. Ou execute: Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Alternativa: Execute qualquer comando .ps1 com bypass inline:
# powershell -ExecutionPolicy Bypass -File .\make.ps1 infra-up

# Para uso permanente, adicione isto ao seu perfil do PowerShell:
# $PROFILE = $PROFILE -eq "" ? $PSHOME\profile.ps1 : $PROFILE
# notepad $PROFILE
# E adicione: Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process

Write-Host "PowerShell Script Setup" -ForegroundColor Cyan
Write-Host "======================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Se recebe erro 'cannot be loaded because running scripts is disabled'" -ForegroundColor Yellow
Write-Host ""
Write-Host "Solucoes:" -ForegroundColor Green
Write-Host ""
Write-Host "1. TEMPORARIA (apenas para esta sessao):" -ForegroundColor Yellow
Write-Host "   Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process" -ForegroundColor White
Write-Host ""
Write-Host "2. PERMANENTE (para este usuario):" -ForegroundColor Yellow
Write-Host "   Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser" -ForegroundColor White
Write-Host ""
Write-Host "3. DIRETO NA LINHA DE COMANDO:" -ForegroundColor Yellow
Write-Host "   powershell -ExecutionPolicy Bypass -File .\make.ps1 infra-up" -ForegroundColor White
Write-Host ""
