# pkg/genproto

Folder ini berisi kode hasil **generate** dari file `proto/book/v1/book.proto`.

Jalankan salah satu:

```bash
make generate          # memakai buf (disarankan)
# atau lihat README bagian "Generate kode" untuk perintah protoc manual
```

Setelah generate, akan muncul:

```
pkg/genproto/book/v1/book.pb.go       # message
pkg/genproto/book/v1/book_grpc.pb.go  # server/client gRPC
pkg/genproto/book/v1/book.pb.gw.go    # REST gateway
docs/swagger/openapi.swagger.json     # OpenAPI hasil generate
```
