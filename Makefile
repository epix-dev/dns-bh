build: prepare export-file hazard malware acl

prepare:
	mkdir -p build/dns-bh_master/etc build/dns-bh_master/bin
	cp config.yml build/dns-bh_master/etc

	mkdir -p build/dns-bh_node/etc build/dns-bh_node/bin
	cp config.yml build/dns-bh_node/etc

	cp -r contrib build/contrib

export-file: lib/lib.go cmd/export-file/main.go
	go build -o build/dns-bh_node/bin/export-file cmd/export-file/main.go

malware: lib/lib.go cmd/malware/main.go
	go build -o build/dns-bh_master/bin/malware cmd/malware/main.go

hazard: lib/lib.go cmd/hazard/main.go
	go build -o build/dns-bh_master/bin/hazard cmd/hazard/main.go

acl: lib/lib.go cmd/acl/main.go
	go build --tags "libsqlite3 linux" -o build/dns-bh_master/bin/acl cmd/acl/main.go

clean:
	go clean
	rm -rf build
