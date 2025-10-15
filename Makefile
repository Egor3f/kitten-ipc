default:
	@echo "Please read Makefile for available targets"

# Very bad. But i had no choice! its internal!
vendor_tsgo:
	@mkdir -p ./kitcom/internal/tsgo
	@git clone --depth 1 https://github.com/microsoft/typescript-go
	@echo Renaming packages...
	@find ./typescript-go/internal -type file -name "*.go" -exec sed -i -e 's!"github.com/microsoft/typescript-go/internal!"efprojects.com/kitten-ipc/kitcom/internal/tsgo!g' {} \;
	@cp -r ./typescript-go/internal/* ./kitcom/internal/tsgo
	@git add ./kitcom/internal/
	@echo Cleaning up...
	@rm -rf @rm -rf typescript-go
	echo Successfully copied tsgo code and renamed packages.

.PHONY: vendor_tsgo
