# Local S3 Console

A React-based web application that provides an AWS S3 console-like interface for managing local S3 buckets and objects. Built with modern technologies including Vite, React, TypeScript, TanStack Router, TanStack Query, TailwindCSS, and AWS SDK v3.

## Features

### Bucket Management
- **List Buckets**: View all local S3 buckets in your account
- **Create Bucket**: Create new buckets with optional tagging support
- **Delete Bucket**: Remove empty buckets
- **Bucket Policy**: View and edit bucket policies (JSON editor)
- **Bucket Notifications**: Configure bucket notification settings

### Object Management
- **List Objects**: Browse objects within a bucket
- **Upload Objects**: Upload files with optional tagging support
- **Delete Objects**: Remove objects from buckets
- **Object Tagging**: Add tags during object upload

### Configuration
- **AWS Credentials**: Configure access key, secret key, region
- **Custom Endpoint**: Support for local S3-compatible services (MinIO, LocalStack)
- **Force Path Style**: Enable path-style URLs for local S3 services

## Tech Stack

- **Vite**: Fast build tool and dev server
- **React 19**: UI framework
- **TypeScript**: Type-safe development
- **TanStack Router**: File-based routing with type safety
- **TanStack Query**: Data fetching and caching
- **TailwindCSS v4**: Utility-first CSS framework
- **AWS SDK v3**: AWS S3 client library

## Getting Started

### Prerequisites

- Node.js 18+
- pnpm (package manager)
- AWS credentials or local S3-compatible service (MinIO, LocalStack)

### Installation

1. Install dependencies:
```bash
pnpm install
```

2. Create a `.env` file in the root directory based on `.env.example`:
```bash
cp .env.example .env
```

3. Configure your environment variables in `.env`:
```env
VITE_AWS_REGION=us-east-1
VITE_AWS_ACCESS_KEY_ID=your-access-key-id
VITE_AWS_SECRET_ACCESS_KEY=your-secret-access-key

# Optional: For local S3 services (MinIO, LocalStack)
VITE_AWS_ENDPOINT=http://localhost:9000
VITE_AWS_FORCE_PATH_STYLE=true
```

### Development

Run the development server:
```bash
pnpm dev
```

The application will be available at http://localhost:5173

### Build

Create a production build:
```bash
pnpm build
```

Preview the production build:
```bash
pnpm preview
```

## Configuration

### Environment Variables

The application is configured using environment variables. All environment variables must be prefixed with `VITE_` to be accessible in the browser.

#### Required Variables

- `VITE_AWS_REGION` - AWS region (e.g., `us-east-1`)
- `VITE_AWS_ACCESS_KEY_ID` - AWS access key ID
- `VITE_AWS_SECRET_ACCESS_KEY` - AWS secret access key

#### Optional Variables

- `VITE_AWS_ENDPOINT` - Custom S3 endpoint URL (for MinIO, LocalStack, etc.)
- `VITE_AWS_FORCE_PATH_STYLE` - Set to `true` to use path-style URLs (required for MinIO/LocalStack)

### Example Configurations

#### AWS S3
```env
VITE_AWS_REGION=us-east-1
VITE_AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
VITE_AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

#### MinIO
```env
VITE_AWS_REGION=us-east-1
VITE_AWS_ACCESS_KEY_ID=minioadmin
VITE_AWS_SECRET_ACCESS_KEY=minioadmin
VITE_AWS_ENDPOINT=http://localhost:9000
VITE_AWS_FORCE_PATH_STYLE=true
```

#### LocalStack
```env
VITE_AWS_REGION=us-east-1
VITE_AWS_ACCESS_KEY_ID=test
VITE_AWS_SECRET_ACCESS_KEY=test
VITE_AWS_ENDPOINT=http://localhost:4566
VITE_AWS_FORCE_PATH_STYLE=true
```

## Project Structure

```
web/
├── src/
│   ├── components/        # Reusable UI components
│   ├── hooks/            # Custom React hooks
│   ├── lib/              # Core libraries and services
│   │   ├── s3Client.ts   # S3 client initialization
│   │   ├── s3Service.ts  # S3 API operations
│   │   └── queryClient.ts # TanStack Query configuration
│   ├── routes/           # File-based routes
│   │   ├── __root.tsx    # Root layout
│   │   ├── index.tsx     # Bucket list page
│   │   └── buckets.$bucketName.tsx # Bucket detail page
│   ├── index.css         # Global styles with Tailwind
│   └── main.tsx          # Application entry point
├── public/               # Static assets
├── vite.config.ts        # Vite configuration
├── tsconfig.json         # TypeScript configuration
└── package.json          # Project dependencies
```

## Available Operations

### Bucket Operations
- listBuckets() - List all buckets
- createBucket(name, tags?) - Create bucket with optional tags
- deleteBucket(name) - Delete bucket
- getBucketTags(name) - Get bucket tags
- getBucketPolicy(name) - Get bucket policy
- putBucketPolicy(name, policy) - Update bucket policy
- getBucketNotificationConfiguration(name) - Get notification config
- putBucketNotificationConfiguration(name, config) - Update notification config

### Object Operations
- listObjects(bucket, prefix?) - List objects in bucket
- putObject(bucket, key, body, tags?) - Upload object with optional tags
- getObject(bucket, key) - Download object
- deleteObject(bucket, key) - Delete object
- getObjectTags(bucket, key) - Get object tags

## Development Notes

### TanStack Router
- Routes are automatically generated from src/routes/ directory
- Route tree is generated to src/routeTree.gen.ts (ignored in git)
- File naming convention: __root.tsx, index.tsx, page.$param.tsx

### TanStack Query
- Configured with 5-minute stale time
- Automatic cache invalidation after mutations
- Query keys: ['buckets'], ['objects', bucketName], etc.

### TailwindCSS v4
- Uses new @tailwindcss/postcss plugin
- No tailwind.config.js needed for basic setup
- Includes autoprefixer for browser compatibility

## Security Notes

- **Environment variables are embedded in the built JavaScript bundle** - They are NOT secret in the browser
- For production use, implement proper backend authentication/authorization
- Consider using AWS Cognito, temporary credentials, or a backend proxy for production deployments
- Never commit `.env` files to version control (already in `.gitignore`)
- The `.env` file should only be used for development/testing with non-sensitive credentials

## License

MIT
