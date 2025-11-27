@echo off
REM Wrapper batch para Makefile - funciona em Windows sem make/gmake

if "%1"=="" (
    call :show_help
    goto end
)

if "%1"=="help" (
    call :show_help
    goto end
)

if "%1"=="localstack-start" (
    echo Starting LocalStack...
    localstack start -d
    timeout /t 10 /nobreak
    localstack status
    goto end
)

if "%1"=="localstack-stop" (
    echo Stopping LocalStack...
    localstack stop
    goto end
)

if "%1"=="localstack-status" (
    echo LocalStack Status:
    localstack status
    goto end
)

if "%1"=="cognito-local-start" (
    echo Starting cognito-local...
    docker-compose -f docker-compose.cognito-local.yaml up -d
    timeout /t 10 /nobreak
    echo cognito-local started at http://localhost:9229
    goto end
)

if "%1"=="cognito-local-stop" (
    echo Stopping cognito-local...
    docker-compose -f docker-compose.cognito-local.yaml down
    goto end
)

if "%1"=="cognito-local-setup" (
    echo Setting up cognito-local...
    pushd infra
    call setup-cognito-local.sh
    popd
    goto end
)

if "%1"=="cognito-local-clean" (
    echo Cleaning cognito-local...
    docker-compose -f docker-compose.cognito-local.yaml down -v
    for /r infra\cognito-local-config %%A in (*.json) do del "%%A"
    goto end
)

if "%1"=="tflocal-init" (
    echo Initializing Terraform Local...
    pushd infra
    ren cognito.tf cognito.tf.skip >nul 2>&1
    tflocal init
    ren cognito.tf.skip cognito.tf >nul 2>&1
    popd
    goto end
)

if "%1"=="tflocal-plan" (
    echo Planning Terraform...
    pushd infra
    ren cognito.tf cognito.tf.skip >nul 2>&1
    tflocal plan -var="use_localstack=true"
    ren cognito.tf.skip cognito.tf >nul 2>&1
    popd
    goto end
)

if "%1"=="tflocal-apply" (
    echo Applying infrastructure...
    pushd infra
    ren cognito.tf cognito.tf.skip >nul 2>&1
    tflocal apply -auto-approve -var="use_localstack=true"
    ren cognito.tf.skip cognito.tf >nul 2>&1
    popd
    goto end
)

if "%1"=="tflocal-destroy" (
    echo Destroying infrastructure...
    pushd infra
    ren cognito.tf cognito.tf.skip >nul 2>&1
    tflocal destroy -auto-approve -var="use_localstack=true"
    ren cognito.tf.skip cognito.tf >nul 2>&1
    popd
    goto end
)

if "%1"=="infra-up" (
    echo Starting infrastructure...
    echo.
    echo 1. Starting LocalStack...
    call make.bat localstack-start
    echo.
    echo 2. Starting cognito-local...
    call make.bat cognito-local-start
    echo.
    echo 3. Initializing Terraform...
    call make.bat tflocal-init
    echo.
    echo 4. Setting up cognito-local...
    call make.bat cognito-local-setup
    echo.
    echo 5. Applying infrastructure...
    call make.bat tflocal-apply
    echo.
    echo Infrastructure started!
    goto end
)

if "%1"=="infra-down" (
    echo Stopping infrastructure...
    echo.
    echo 1. Destroying Terraform...
    call make.bat tflocal-destroy
    echo.
    echo 2. Stopping cognito-local...
    call make.bat cognito-local-clean
    echo.
    echo 3. Stopping LocalStack...
    call make.bat localstack-stop
    echo.
    echo Infrastructure stopped!
    goto end
)

echo Unknown target: %1
call :show_help
goto end

:show_help
echo ========================================
echo Const-Software Make Commands
echo ========================================
echo.
echo LocalStack:
echo   make.bat localstack-start
echo   make.bat localstack-stop
echo   make.bat localstack-status
echo.
echo Cognito-Local:
echo   make.bat cognito-local-start
echo   make.bat cognito-local-stop
echo   make.bat cognito-local-setup
echo.
echo Terraform:
echo   make.bat tflocal-init
echo   make.bat tflocal-plan
echo   make.bat tflocal-apply
echo   make.bat tflocal-destroy
echo.
echo Combined:
echo   make.bat infra-up
echo   make.bat infra-down
echo.
goto end

:end
