# File Upload System

This application includes a comprehensive file upload system that automatically uses AWS S3 when configured, or falls back to database storage when S3 is not available.

## Features

- **Dual Storage Backend**: Automatically uses S3 when AWS credentials are available, otherwise stores files in PostgreSQL database
- **Secure Upload**: File uploads require authentication
- **File Management**: Complete CRUD operations for uploaded files
- **Public Downloads**: File downloads are publicly accessible (no auth required)
- **Metadata Storage**: All file metadata is stored in the database regardless of storage backend
- **Rate Limiting**: Built-in protection against abuse
- **File Size Limits**: Configurable file size limits (default 10MB)
- **Content Type Detection**: Automatic MIME type detection and storage

## API Endpoints

### File Upload

```
POST /api/files/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>
```

Upload a file. The file should be sent as `file` field in the multipart form.

### File Download

```
GET /api/files/{id}/download
```

Download a file by its ID. For S3 files, redirects to the S3 URL. For database files, serves the content directly.

### Get File Information

```
GET /api/files/{id}
```

Get metadata for a specific file.

### Get File URL

```
GET /api/files/{id}/url
```

Get the URL for accessing a file (useful for frontend applications).

### List Files

```
GET /api/files?limit=10&offset=0
Authorization: Bearer <token>
```

List uploaded files with pagination. Requires authentication.

### Delete File

```
DELETE /api/files/{id}
Authorization: Bearer <token>
```

Delete a file from both storage backend and database.

### Storage Status

```
GET /api/files/storage/status
```

Get information about the current storage configuration (public endpoint).

## Configuration

### Environment Variables

Add these to your `.env` file:

```bash
# AWS S3 Configuration (Optional - for file storage)
# If these are not set, files will be stored in the database
AWS_ACCESS_KEY_ID=your-aws-access-key-id
AWS_SECRET_ACCESS_KEY=your-aws-secret-access-key
AWS_REGION=us-east-1
AWS_S3_BUCKET=your-s3-bucket-name

# File Upload Configuration
MAX_FILE_SIZE_MB=10
```

### AWS S3 Setup

1. **Create an S3 Bucket**:

   - Go to AWS S3 Console
   - Create a new bucket
   - Note the bucket name and region

2. **Create IAM User**:

   - Go to AWS IAM Console
   - Create a new user with programmatic access
   - Attach the following policy:

   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"],
         "Resource": "arn:aws:s3:::your-bucket-name/*"
       }
     ]
   }
   ```

3. **Configure Environment Variables**:
   ```bash
   AWS_ACCESS_KEY_ID=your-access-key
   AWS_SECRET_ACCESS_KEY=your-secret-key
   AWS_REGION=us-east-1
   AWS_S3_BUCKET=your-bucket-name
   ```

### Database Storage (Fallback)

If AWS S3 credentials are not provided, files will be stored directly in the PostgreSQL database as BLOB data. This is suitable for:

- Development environments
- Small-scale applications
- Applications with limited file storage needs

## Usage Examples

### Upload a File (JavaScript)

```javascript
const formData = new FormData();
formData.append('file', fileInput.files[0]);

const response = await fetch('/api/files/upload', {
  method: 'POST',
  headers: {
    Authorization: `Bearer ${token}`,
  },
  body: formData,
});

const result = await response.json();
console.log('Uploaded file ID:', result.data.id);
```

### Download a File

```javascript
// Get file URL first
const urlResponse = await fetch(`/api/files/${fileId}/url`);
const urlData = await urlResponse.json();

// Then download
window.open(urlData.data, '_blank');
```

### List Files

```javascript
const response = await fetch('/api/files?limit=20&offset=0', {
  headers: {
    Authorization: `Bearer ${token}`,
  },
});

const files = await response.json();
console.log('Files:', files.data);
```

## Storage Backend Selection Logic

The system automatically chooses the storage backend based on configuration:

1. **S3 Storage** (Preferred):

   - Used when all AWS credentials are provided
   - Files are uploaded to S3
   - Metadata stored in database
   - File content not stored in database (only S3 key)

2. **Database Storage** (Fallback):
   - Used when AWS credentials are missing or invalid
   - Files stored as BLOB in PostgreSQL
   - Metadata and content both stored in database
   - Suitable for development/small scale use

## Security Considerations

- **Authentication**: File uploads require JWT authentication
- **File Size Limits**: 10MB default limit (configurable)
- **Content Type Validation**: Files are accepted regardless of type (consider adding validation)
- **S3 Permissions**: Minimal IAM permissions for security
- **Rate Limiting**: Built-in protection against abuse

## Database Schema

The `files` table includes:

```sql
CREATE TABLE files (
  id SERIAL PRIMARY KEY,
  created_at VARCHAR(255),
  updated_at VARCHAR(255),
  file_name VARCHAR(255) NOT NULL,
  content_type VARCHAR(255),
  file_size BIGINT,
  location VARCHAR(255),  -- S3 key or database identifier
  content BYTEA,         -- File content (database storage only)
  storage_type VARCHAR(255) DEFAULT 'database'
);
```

## Performance Considerations

- **S3 Storage**: Better for large files and high traffic
- **Database Storage**: Better for small files and low traffic
- **CDN Integration**: Consider using CloudFront with S3 for better performance
- **File Processing**: Add image resizing, compression, or virus scanning as needed

## Monitoring and Maintenance

- **Storage Usage**: Monitor S3 bucket size and database storage
- **File Cleanup**: Implement cleanup policies for temporary files
- **Backup Strategy**: Ensure S3 files are backed up appropriately
- **Access Logs**: Monitor file access patterns

## Troubleshooting

### Common Issues

1. **S3 Upload Fails**:

   - Check AWS credentials
   - Verify S3 bucket permissions
   - Check bucket region configuration

2. **Database Storage Issues**:

   - Check PostgreSQL connection
   - Verify BYTEA column size limits
   - Monitor database storage space

3. **File Download Issues**:
   - Check file permissions
   - Verify file exists in storage
   - Check S3 bucket policies

### Logs

Check application logs for detailed error information:

```bash
# View recent logs
tail -f /var/log/your-app.log

# Search for file-related errors
grep "file" /var/log/your-app.log
```

## Future Enhancements

- **File Processing**: Image resizing, compression
- **Virus Scanning**: Integrate with antivirus services
- **CDN Integration**: CloudFront or similar
- **File Versioning**: Keep multiple versions of files
- **Batch Operations**: Upload/download multiple files
- **Storage Migration**: Migrate files between storage backends



