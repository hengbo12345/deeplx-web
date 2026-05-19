# DeepLX Web

A modern web interface for DeepLX - providing free translation services for text and Word documents.

## Features

- **Text Translation**: Translate text between 28+ languages with auto-detection
- **Document Translation**: Upload and translate Word documents (.docx, .doc) while preserving formatting
- **Modern UI**: Clean, responsive interface built with React and Tailwind CSS
- **File Upload Support**: Handles documents up to 10MB
- **Real-time Translation**: Fast translation powered by DeepLX
- **Free & Open Source**: No API keys required, completely free to use

## Supported Languages

Bulgarian, Chinese, Czech, Danish, Dutch, English, Estonian, Finnish, French, German, Greek, Hungarian, Indonesian, Italian, Japanese, Korean, Lithuanian, Latvian, Norwegian, Polish, Portuguese, Romanian, Russian, Slovak, Slovenian, Spanish, Swedish, Turkish, Ukrainian

## Tech Stack

### Backend
- Go 1.21
- Chi router
- Zap logging
- Lumberjack for log rotation
- DOCX handling for document translation

### Frontend
- React 18
- TypeScript
- Vite
- Tailwind CSS
- React Router

## Project Structure

```
.
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # Server entry point
│   ├── internal/
│   │   ├── config/              # Configuration management
│   │   ├── handler/             # HTTP handlers
│   │   ├── middleware/          # HTTP middleware
│   │   ├── models/              # Data models
│   │   └── service/             # Business logic
│   ├── pkg/
│   │   └── utils/               # Utility functions
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/          # React components
│   │   ├── pages/               # Page components
│   │   ├── lib/                 # Utilities and API client
│   │   └── styles/              # Global styles
│   ├── package.json
│   ├── vite.config.ts
│   └── tailwind.config.js
├── docker-compose.yml
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Docker and Docker Compose (optional)

### Deployment Options

#### Option 1: Docker Compose (Development/Test)

1. Clone the repository:
```bash
git clone <repository-url>
cd sayHello
```

2. Configure environment variables:
```bash
cp .env.example .env
# Edit .env and set DEEPLX_TOKEN
```

3. Start the services:
```bash
docker-compose up -d
```

4. Access the application:
- Frontend: http://localhost:9449
- Backend API: http://localhost:8449

#### Option 2: Production HTTPS Deployment

The project is pre-configured for HTTPS production deployment (domain: deeplxweb.yourdomain.com).

1. Prepare SSL certificates:
```bash
# Place SSL certificate files in frontend/ssl/ directory
cp your-cert.cer frontend/ssl/deeplxweb.cer
cp your-cert.key frontend/ssl/deeplxweb.key
```

2. Configure environment variables:
```bash
cp .env.example .env
# Edit .env and set DEEPLX_TOKEN
```

3. Update nginx.conf with your domain (if needed):
```nginx
server_name  deeplxweb.yourdoamin.com;  # Change to your domain
```

4. Build and start:
```bash
docker-compose up -d
```

5. Access: https://deeplxweb.yourdoamin.com (or your configured domain)

**Note**: The production configuration uses nginx to proxy requests to backend on 127.0.0.1:8449. Ensure the backend service is running on that port.

### Manual Setup

#### Backend

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file (optional):
```env
SERVER_PORT=9448
DEEPLX_URL=http://localhost:1188
DEEPLX_TOKEN=
AUTH_TOKEN=
UPLOAD_PATH=./uploads
FILE_MAX_AGE=24h
UPLOAD_MAX_SIZE=10485760
CLEANUP_INTERVAL=1h
LOG_LEVEL=info
```

4. Run the server:
```bash
go run cmd/server/main.go
```

The backend will start on port 9448 by default.

#### Frontend

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

The frontend will start on port 3000.

### Building for Production

#### Backend
```bash
cd backend
go build -o deeplx-web ./cmd/server
```

#### Frontend
```bash
cd frontend
npm run build
```

The built files will be in the `frontend/dist` directory.

## API Endpoints

### Health Check
```
GET /health
```

### Text Translation
```
POST /api/translate
Content-Type: application/json

{
  "text": "Hello, world!",
  "source_lang": "EN",
  "target_lang": "ZH"
}
```

### Document Translation
```
POST /api/translate/document
Content-Type: multipart/form-data

file: <document>
source_lang: EN
target_lang: ZH
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| SERVER_HOST | Backend bind host | 127.0.0.1 |
| SERVER_PORT | Backend internal port | 9448 |
| BACKEND_PORT | Backend exposed port | 8449 |
| FRONTEND_PORT | Frontend HTTPS port | 9449 |
| DEEPLX_URL | DeepLX service URL | http://localhost:1188 |
| DEEPLX_TOKEN | DeepLX authentication token | (empty) |
| UPLOAD_PATH | File upload directory | ./uploads |
| FILE_MAX_AGE | File retention period | 24h |
| UPLOAD_MAX_SIZE | Maximum file size | 10485760 (10MB) |
| CLEANUP_INTERVAL | Cleanup interval | 1h |
| LOG_LEVEL | Logging level | info |

### Production HTTPS Setup

The project includes production-ready HTTPS configuration for `deeplxweb.yourdoamin.com`. To use your own domain:

1. Update `frontend/nginx.conf`:
   - Change `server_name deeplxweb.yourdoamin.com` to your domain
   - Update SSL certificate paths if needed

2. Place SSL certificates in `frontend/ssl/`:
   - `deeplxweb.cer` (or your domain certificate)
   - `deeplxweb.key` (or your domain private key)

3. Update `docker-compose.yml` to mount SSL directory:
   ```yaml
   frontend:
     volumes:
       - ./frontend/ssl:/etc/nginx/conf.d:ro
   ```

## Development

### Backend Development
```bash
cd backend
go run cmd/server/main.go
```

### Frontend Development
```bash
cd frontend
npm run dev
```

### Linting

Frontend:
```bash
npm run lint
```

## Docker

### Build Backend Image
```bash
docker build -t deeplx-web ./backend
```

### Run with Docker
```bash
docker run -p 9448:9448 \
  -e DEEPLX_URL=http://your-deeplx-service:1188 \
  deeplx-web
```

## License

This project is open source and available under the MIT License.

## Acknowledgments

- [DeepLX](https://github.com/OwO-Network/DeepLX) - The unofficial open-source DeepL translation API
- [Go Chi](https://github.com/go-chi/chi) - Lightweight, idiomatic and composable router for building Go HTTP services
- [React](https://react.dev/) - A JavaScript library for building user interfaces
- [Tailwind CSS](https://tailwindcss.com/) - A utility-first CSS framework

## Support

For issues, questions, or contributions, please visit the project repository.
