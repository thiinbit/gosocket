#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o gosocket .
GOOS=linux GOARCH=386 go build -o gosocket_linux_386 .
GOOS=linux GOARCH=arm64 go build -o gosocket_linux_arm64 .
GOOS=linux GOARCH=arm GOARM=7 go build -o gosocket_linux_arm7 .
GOOS=linux GOARCH=arm GOARM=6 go build -o gosocket_linux_arm6 .
GOOS=linux GOARCH=arm GOARM=5 go build -o gosocket_linux_arm5 .
GOOS=linux GOARCH=mips go build -o gosocket_linux_mips .
GOOS=linux GOARCH=mipsle go build -o gosocket_linux_mipsle .
GOOS=linux GOARCH=mips GOMIPS=softfloat go build -o gosocket_linux_mips_softfloat .
GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -o gosocket_linux_mipsle_softfloat .
GOOS=linux GOARCH=mips64 go build -o gosocket_linux_mips64 .
GOOS=linux GOARCH=mips64le go build -o gosocket_linux_mips64le .
GOOS=linux GOARCH=mips64 GOMIPS=softfloat go build -o gosocket_linux_mips64_softfloat .
GOOS=linux GOARCH=mips64le GOMIPS=softfloat go build -o gosocket_linux_mips64le_softfloat .
GOOS=linux GOARCH=ppc64 go build -o gosocket_linux_ppc64 .
GOOS=linux GOARCH=ppc64le go build -o gosocket_linux_ppc64le .
GOOS=darwin GOARCH=amd64 go build -o gosocket_darwin_amd64 .
GOOS=windows GOARCH=amd64 go build -o gosocket_windows_amd64.exe .
GOOS=windows GOARCH=386 go build -o gosocket_windows_386.exe .