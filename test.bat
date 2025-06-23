@echo off
echo Running tests for saurfang project...
echo.

echo Running all tests...
go test ./... -v

echo.
echo Running specific handler tests...
go test ./internal/handler/taskhandler -v

echo.
echo Tests completed.
pause 