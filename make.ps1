#!/usr/bin/env pwsh
# PowerShell wrapper para Makefile

param([string]$Target = "help")

switch ($Target) {
    "help" {
        Write-Host "========================================" -ForegroundColor Cyan
        Write-Host "Const-Software Make Commands" -ForegroundColor Cyan
        Write-Host "========================================" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "LocalStack:" -ForegroundColor Yellow
        Write-Host "  .\make.ps1 localstack-start"
        Write-Host "  .\make.ps1 localstack-stop"
        Write-Host "  .\make.ps1 localstack-status"
        Write-Host ""
        Write-Host "Cognito-Local:" -ForegroundColor Yellow
        Write-Host "  .\make.ps1 cognito-local-start"
        Write-Host "  .\make.ps1 cognito-local-stop"
        Write-Host "  .\make.ps1 cognito-local-setup"
        Write-Host ""
        Write-Host "Combined:" -ForegroundColor Yellow
        Write-Host "  .\make.ps1 infra-up"
        Write-Host "  .\make.ps1 infra-down"
        Write-Host ""
    }
    
    "localstack-start" {
        Write-Host "Starting LocalStack..." -ForegroundColor Green
        & localstack start -d
        Start-Sleep -Seconds 10
        & localstack status
    }
    
    "localstack-stop" {
        Write-Host "Stopping LocalStack..." -ForegroundColor Red
        & localstack stop
    }
    
    "localstack-status" {
        Write-Host "LocalStack Status:" -ForegroundColor Cyan
        try {
            & localstack status
        } catch {
            Write-Host "LocalStack is not running" -ForegroundColor Red
        }
    }
    
    "cognito-local-start" {
        Write-Host "Starting cognito-local..." -ForegroundColor Green
        & docker-compose -f docker-compose.cognito-local.yaml up -d
        Start-Sleep -Seconds 10
        Write-Host "cognito-local started at http://localhost:9229" -ForegroundColor Green
    }
    
    "cognito-local-stop" {
        Write-Host "Stopping cognito-local..." -ForegroundColor Red
        & docker-compose -f docker-compose.cognito-local.yaml down
    }
    
    "cognito-local-setup" {
        Write-Host "Setting up cognito-local..." -ForegroundColor Green
        Push-Location infra
        & ./setup-cognito-local.sh
        Pop-Location
    }
    
    "cognito-local-clean" {
        Write-Host "Cleaning cognito-local..." -ForegroundColor Yellow
        & docker-compose -f docker-compose.cognito-local.yaml down -v
        Remove-Item -Path infra\cognito-local-config\*.json -Force -ErrorAction SilentlyContinue
    }
    
    "tflocal-init" {
        Write-Host "Initializing Terraform Local..." -ForegroundColor Green
        Push-Location infra
        Rename-Item -Path cognito.tf -NewName cognito.tf.skip -Force -ErrorAction SilentlyContinue
        & tflocal init
        Rename-Item -Path cognito.tf.skip -NewName cognito.tf -Force -ErrorAction SilentlyContinue
        Pop-Location
    }
    
    "tflocal-plan" {
        Write-Host "Planning Terraform..." -ForegroundColor Cyan
        Push-Location infra
        Rename-Item -Path cognito.tf -NewName cognito.tf.skip -Force -ErrorAction SilentlyContinue
        & tflocal plan -var="use_localstack=true"
        Rename-Item -Path cognito.tf.skip -NewName cognito.tf -Force -ErrorAction SilentlyContinue
        Pop-Location
    }
    
    "tflocal-apply" {
        Write-Host "Applying infrastructure..." -ForegroundColor Green
        Push-Location infra
        Rename-Item -Path cognito.tf -NewName cognito.tf.skip -Force -ErrorAction SilentlyContinue
        & tflocal apply -auto-approve -var="use_localstack=true"
        Rename-Item -Path cognito.tf.skip -NewName cognito.tf -Force -ErrorAction SilentlyContinue
        Pop-Location
    }
    
    "tflocal-destroy" {
        Write-Host "Destroying infrastructure..." -ForegroundColor Red
        Push-Location infra
        Rename-Item -Path cognito.tf -NewName cognito.tf.skip -Force -ErrorAction SilentlyContinue
        & tflocal destroy -auto-approve -var="use_localstack=true"
        Rename-Item -Path cognito.tf.skip -NewName cognito.tf -Force -ErrorAction SilentlyContinue
        Pop-Location
    }
    
    "infra-up" {
        Write-Host "Starting infrastructure..." -ForegroundColor Green
        Write-Host ""
        Write-Host "1. Starting LocalStack..." -ForegroundColor Cyan
        & .\make.ps1 localstack-start
        Write-Host ""
        Write-Host "2. Starting cognito-local..." -ForegroundColor Cyan
        & .\make.ps1 cognito-local-start
        Write-Host ""
        Write-Host "3. Initializing Terraform..." -ForegroundColor Cyan
        & .\make.ps1 tflocal-init
        Write-Host ""
        Write-Host "4. Setting up cognito-local..." -ForegroundColor Cyan
        & .\make.ps1 cognito-local-setup
        Write-Host ""
        Write-Host "5. Applying infrastructure..." -ForegroundColor Cyan
        & .\make.ps1 tflocal-apply
        Write-Host ""
        Write-Host "Infrastructure started!" -ForegroundColor Green
    }
    
    "infra-down" {
        Write-Host "Stopping infrastructure..." -ForegroundColor Red
        Write-Host ""
        Write-Host "1. Destroying Terraform..." -ForegroundColor Cyan
        & .\make.ps1 tflocal-destroy
        Write-Host ""
        Write-Host "2. Stopping cognito-local..." -ForegroundColor Cyan
        & .\make.ps1 cognito-local-clean
        Write-Host ""
        Write-Host "3. Stopping LocalStack..." -ForegroundColor Cyan
        & .\make.ps1 localstack-stop
        Write-Host ""
        Write-Host "Infrastructure stopped!" -ForegroundColor Green
    }
    
    default {
        Write-Host "Unknown target: $Target" -ForegroundColor Red
        & .\make.ps1 help
    }
}
