AUTHOR			= Adam Kubica <xcdr@kaizen-step.com>
BUILD_VERSION	= 0.1.4-beta
BUILD_BRANCH	= $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE		= $(shell date +%Y%m%d%H%M)

BUILD_DIR		= build

# LDFLAGS			:= -extldflags '-static'
LDFLAGS			+= -X 'main.author=${AUTHOR}'
LDFLAGS 		+= -X 'main.version=${BUILD_VERSION}'
LDFLAGS 		+= -X 'main.build=${BUILD_DATE}.${BUILD_BRANCH}'

.prepare:
	mkdir -p ${BUILD_DIR}/dns-bh_master/etc ${BUILD_DIR}/dns-bh_master/bin
	cp config.yml ${BUILD_DIR}/dns-bh_master/etc

	mkdir -p ${BUILD_DIR}/dns-bh_node/etc ${BUILD_DIR}/dns-bh_node/bin
	cp config.yml ${BUILD_DIR}/dns-bh_node/etc

	cp -r contrib ${BUILD_DIR}/

.export-file: cmd/export-file/main.go
	GOFLAGS=-mod=vendor CGO_ENABLE=0 \
	go build -a -installsuffix cgo -ldflags "${LDFLAGS}" \
	-o ${BUILD_DIR}/dns-bh_node/bin/export-file cmd/export-file/main.go

.malware: cmd/malware/main.go
	GOFLAGS=-mod=vendor CGO_ENABLE=0 \
	go build -a -installsuffix cgo -ldflags "${LDFLAGS}" \
	-o ${BUILD_DIR}/dns-bh_master/bin/malware cmd/malware/main.go

.hazard: cmd/hazard/main.go
	GOFLAGS=-mod=vendor CGO_ENABLE=0 \
	go build -a -installsuffix cgo -ldflags "${LDFLAGS}" \
	-o ${BUILD_DIR}/dns-bh_master/bin/hazard cmd/hazard/main.go

.cert_hole: cmd/cert_hole/main.go
	GOFLAGS=-mod=vendor CGO_ENABLE=0 \
	go build -a -installsuffix cgo -ldflags "${LDFLAGS}" \
	-o ${BUILD_DIR}/dns-bh_master/bin/cert_hole cmd/cert_hole/main.go

.acl: cmd/acl/main.go
	GOFLAGS=-mod=vendor CGO_ENABLE=0 \
	go build -a -installsuffix cgo -ldflags "${LDFLAGS}" --tags "libsqlite3 linux" \
	-o ${BUILD_DIR}/dns-bh_master/bin/acl cmd/acl/main.go

build: .prepare .export-file .hazard .malware .acl .cert_hole

tar-file: build
	tar czf dns-bh-${BUILD_VERSION}-bin.tar.gz -C build .

clean:
	GOFLAGS=-mod=vendor go clean
	rm -rf ${BUILD_DIR}
