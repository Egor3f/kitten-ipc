
SHELL := /bin/bash
tsgo_dir = ./internal/tsgo
my_package = efprojects.com/kitten-ipc/kitcom/internal/tsgo


default:
	@echo "Please read Makefile for available targets"


vendor_tsgo:
	@mkdir -p $(tsgo_dir)
	@git clone --depth 1 https://github.com/microsoft/typescript-go
	@find ./typescript-go/internal -type file -name "*.go" -exec sed -i -e 's!"github.com/microsoft/typescript-go/internal!"$(my_package)!g' {} \;
	@cp -r ./typescript-go/internal/* $(tsgo_dir)
	@rm -rf @rm -rf typescript-go


remove_tsgo_tests:
	@find $(tsgo_dir) -name "*_test.go" -exec rm {} \;


# tree shaking. written in make just for "fun"
# caution: may cause eye hazard
remove_tsgo_unused:
	@set -e ; \
	dirs=`find $(tsgo_dir) -type d -mindepth 1 -maxdepth 1` ; \
	nessesary_old="parser " ; \
	nessesary="$$nessesary_old" ; \
	while true; do \
		for d in $$dirs; do \
			pkg=`basename "$$d"` ; \
			for usedIn in $$nessesary; do \
				if grep -q -R "$(my_package)/$$pkg" "$(tsgo_dir)/$$usedIn" > /dev/null; then \
					if [[ "$$nessesary" != *"$$pkg "* ]]; then \
						nessesary="$$nessesary $$pkg " ; \
					fi ; \
					break ; \
				fi ; \
			done ; \
		done ; \
		if [[ "$$nessesary" == "$$nessesary_old" ]]; then \
			break ; \
		fi ; \
		nessesary_old="$$nessesary" ; \
	done ; \
	for d in $$dirs; do \
		pkg=`basename $$d` ; \
		if [[ "$$nessesary" != *"$$pkg "* ]]; then \
		  	echo "removing $$pkg" ; \
		  	rm -rf $(tsgo_dir)/$$pkg ; \
		fi ; \
	done


.PHONY: vendor_tsgo remove_tsgo_tests remove_tsgo_unused
