all:
#	@cd ~/Development/trueblocks-core/build ; make ; cd - 2>/dev/null
	@go mod tidy && go build main.go && mv main ../bin/compare
	@echo "Done..."
