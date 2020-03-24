$CMD=$args[0]
$CURRENT_DIR=Get-Location

docker run --rm `
    -v ${CURRENT_DIR}:/go/src/build `
    -w /go/src/build golang:1.14 sh -c "apt-get update; apt-get install libsqlite3-dev; make $CMD"
