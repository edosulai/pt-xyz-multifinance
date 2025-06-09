$ErrorActionPreference = "Stop"

Write-Host "Generating protobuf code..."

# Create directories
New-Item -ItemType Directory -Force -Path "proto/gen/go" | Out-Null
New-Item -ItemType Directory -Force -Path "proto/gen/openapiv2" | Out-Null

# Clean old files
Remove-Item -Path "proto/gen/go/*" -Recurse -Force -ErrorAction SilentlyContinue
Remove-Item -Path "proto/*.pb.*" -Force -ErrorAction SilentlyContinue

# Generate protobuf code
$cmd = "protoc",
       "-I",".",
       "-I","third_party",
       "--go_out=.",
       "--go_opt=module=github.com/pt-xyz-multifinance",
       "--go-grpc_out=.",
       "--go-grpc_opt=module=github.com/pt-xyz-multifinance",
       "--grpc-gateway_out=.",
       "--grpc-gateway_opt=module=github.com/pt-xyz-multifinance",
       "--openapiv2_out=./proto/gen/openapiv2",
       "proto/user.proto","proto/loan.proto"

& $cmd[0] $cmd[1..($cmd.Length-1)]

Write-Host "Code generation completed successfully"
