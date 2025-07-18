@echo off
:: 设置项目路径和输出目录
::set PROJECT_PATH=C:\path\to\your\project
::set OUTPUT_DIR=%PROJECT_PATH%\bin

:: 定义版本信息
set VERSION=1.0.0

:: 获取当前使用的Go版本
for /f "tokens=3" %%a in ('go version') do set GO_VERSION=%%a

:: 设置构建时间
set BUILD_TIME=%date%_%time%

:: 为了确保构建时间格式一致，去除冒号和空格
set BUILD_TIME=%BUILD_TIME::=-%
set BUILD_TIME=%BUILD_TIME: =_%

:: 切换到项目目录
::cd /d %PROJECT_PATH%

:: 编译命令，嵌入版本信息
go build -o saurfang -ldflags "-X 'main.BuildVersion=%VERSION%' -X 'main.BuildGoVersion=%GO_VERSION%' -X 'main.BuildTime=%BUILD_TIME%'" -p 1

echo Build completed.
pause