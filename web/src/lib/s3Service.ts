import {
  ListBucketsCommand,
  CreateBucketCommand,
  ListObjectsV2Command,
  PutObjectCommand,
  GetObjectCommand,
  HeadObjectCommand,
  PutBucketTaggingCommand,
  GetBucketTaggingCommand,
  PutObjectTaggingCommand,
  GetObjectTaggingCommand,
  PutBucketNotificationConfigurationCommand,
  GetBucketNotificationConfigurationCommand,
  PutBucketPolicyCommand,
  GetBucketPolicyCommand,
  DeleteObjectCommand,
  DeleteBucketCommand,
  type Tag,
  type Bucket,
  type _Object,
} from '@aws-sdk/client-s3'
import { getS3Client } from './s3Client'

export type { Tag }
export type BucketInfo = Bucket
export type S3Object = _Object

// List all buckets
export const listBuckets = async (): Promise<BucketInfo[]> => {
  const client = getS3Client()
  const command = new ListBucketsCommand({})
  const response = await client.send(command)
  return response.Buckets || []
}

// Create bucket with optional tags
export const createBucket = async (bucketName: string, tags?: Tag[]): Promise<void> => {
  const client = getS3Client()
  const createCommand = new CreateBucketCommand({ Bucket: bucketName })
  await client.send(createCommand)

  if (tags && tags.length > 0) {
    const tagCommand = new PutBucketTaggingCommand({
      Bucket: bucketName,
      Tagging: {
        TagSet: tags,
      },
    })
    await client.send(tagCommand)
  }
}

// Get bucket tags
export const getBucketTags = async (bucketName: string): Promise<Tag[]> => {
  const client = getS3Client()
  const command = new GetBucketTaggingCommand({ Bucket: bucketName })
  try {
    const response = await client.send(command)
    return response.TagSet || []
  } catch (error: any) {
    if (error.name === 'NoSuchTagSet') {
      return []
    }
    throw error
  }
}

export interface ListObjectsResult {
  objects: S3Object[]
  commonPrefixes: string[]
}

// List objects in a bucket with delimiter support
export const listObjects = async (
  bucketName: string,
  prefix?: string,
  delimiter?: string
): Promise<ListObjectsResult> => {
  const client = getS3Client()
  const command = new ListObjectsV2Command({
    Bucket: bucketName,
    Prefix: prefix,
    Delimiter: delimiter,
  })
  const response = await client.send(command)
  return {
    objects: response.Contents || [],
    commonPrefixes: response.CommonPrefixes?.map(cp => cp.Prefix || '') || [],
  }
}

// Upload object with optional tags
export const putObject = async (
  bucketName: string,
  key: string,
  body: File | Blob | string,
  tags?: Tag[]
): Promise<void> => {
  const client = getS3Client()

  // Convert File/Blob to ArrayBuffer for browser compatibility
  let bodyContent: ArrayBuffer | string
  if (body instanceof File || body instanceof Blob) {
    bodyContent = await body.arrayBuffer()
  } else {
    bodyContent = body
  }

  const putCommand = new PutObjectCommand({
    Bucket: bucketName,
    Key: key,
    Body: bodyContent,
  })
  await client.send(putCommand)

  if (tags && tags.length > 0) {
    const tagCommand = new PutObjectTaggingCommand({
      Bucket: bucketName,
      Key: key,
      Tagging: {
        TagSet: tags,
      },
    })
    await client.send(tagCommand)
  }
}

// Get object
export const getObject = async (bucketName: string, key: string): Promise<Blob> => {
  const client = getS3Client()
  const command = new GetObjectCommand({
    Bucket: bucketName,
    Key: key,
  })
  const response = await client.send(command)

  if (!response.Body) {
    throw new Error('No body in response')
  }

  // Convert stream to blob
  const stream = response.Body as ReadableStream
  const reader = stream.getReader()
  const chunks: BlobPart[] = []

  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    chunks.push(value)
  }

  return new Blob(chunks)
}

// Get object metadata using HeadObject
export const headObject = async (bucketName: string, key: string) => {
  const client = getS3Client()
  const command = new HeadObjectCommand({
    Bucket: bucketName,
    Key: key,
  })
  const response = await client.send(command)
  return response
}

// Get object tags
export const getObjectTags = async (bucketName: string, key: string): Promise<Tag[]> => {
  const client = getS3Client()
  const command = new GetObjectTaggingCommand({
    Bucket: bucketName,
    Key: key,
  })
  const response = await client.send(command)
  return response.TagSet || []
}

// Put bucket notification configuration
export const putBucketNotificationConfiguration = async (
  bucketName: string,
  configuration: any
): Promise<void> => {
  const client = getS3Client()
  const command = new PutBucketNotificationConfigurationCommand({
    Bucket: bucketName,
    NotificationConfiguration: configuration,
  })
  await client.send(command)
}

// Get bucket notification configuration
export const getBucketNotificationConfiguration = async (
  bucketName: string
): Promise<any> => {
  const client = getS3Client()
  const command = new GetBucketNotificationConfigurationCommand({
    Bucket: bucketName,
  })
  const response = await client.send(command)
  return response
}

// Put bucket policy
export const putBucketPolicy = async (
  bucketName: string,
  policy: string
): Promise<void> => {
  const client = getS3Client()
  const command = new PutBucketPolicyCommand({
    Bucket: bucketName,
    Policy: policy,
  })
  await client.send(command)
}

// Get bucket policy
export const getBucketPolicy = async (bucketName: string): Promise<string> => {
  const client = getS3Client()
  const command = new GetBucketPolicyCommand({
    Bucket: bucketName,
  })
  try {
    const response = await client.send(command)
    return response.Policy || ''
  } catch (error: any) {
    if (error.name === 'NoSuchBucketPolicy') {
      return ''
    }
    throw error
  }
}

// Delete object
export const deleteObject = async (bucketName: string, key: string): Promise<void> => {
  const client = getS3Client()
  const command = new DeleteObjectCommand({
    Bucket: bucketName,
    Key: key,
  })
  await client.send(command)
}

// Delete bucket
export const deleteBucket = async (bucketName: string): Promise<void> => {
  const client = getS3Client()
  const command = new DeleteBucketCommand({
    Bucket: bucketName,
  })
  await client.send(command)
}
