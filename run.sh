_MINIO_DEBUG_NO_EXIT="debug" dlv debug main.go -- server ./database

# 验证
break cmd/auth-handler.go:616

# listobjects
break cmd/object-multipart-handlers.go:1074

break cmd/handler-api.go:319

break cmd/bucket-listobjects-handlers.go:288

ossutil -e 127.0.0.1:9000 -i 12345678 -k 12345678 ls oss://wdg1/
