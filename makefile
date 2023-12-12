all:
	@echo Building...
	@go build *.go
#	@ls -l
	@mkdir -p bin && mv compare bin/compare
	@ls -l bin
#	@cd ~/Development/trueblocks-core/build ; make ; cd -

compress:
	@tar -cvf data.tar store/*
	@gzip data.tar
