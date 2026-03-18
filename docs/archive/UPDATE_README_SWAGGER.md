# README Update - Add Swagger Section

Add this section to README.md after "API Endpoints" section:

---

## 📚 Interactive API Documentation (Swagger)

### Access Swagger UI

Once the server is running, access interactive API documentation:

```
http://localhost:8080/swagger/index.html
```

### Features

- **Interactive Testing** - Try all endpoints directly in browser
- **Request Examples** - Pre-filled with example values
- **Response Schemas** - Complete model documentation
- **cURL Generator** - Auto-generate cURL commands
- **Export** - Download OpenAPI spec (JSON/YAML)
- **Postman Ready** - Import swagger.json to Postman

### Swagger Endpoints

```bash
# Swagger UI (interactive)
http://localhost:8080/swagger/index.html

# OpenAPI JSON spec
http://localhost:8080/swagger/doc.json

# OpenAPI YAML (static file)
docs/swagger.yaml
```

### Screenshot

![Swagger UI](ss/swagger.png)

### Regenerate Documentation

If you modify API annotations:

```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# Rebuild
go build -o idlix-api cmd/api/*.go
```

---

