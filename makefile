all:
#	@cd ~/Development/trueblocks-core/build ; make ; cd - 2>/dev/null
	@cd src && go mod tidy && go build main.go && mv main ../bin/compare && cd - 2>/dev/null
	@echo "Done..."
