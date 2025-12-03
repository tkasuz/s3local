import { S3Client } from '@aws-sdk/client-s3'

let s3Client: S3Client | null = null

const getEnvVar = (key: string): string => {
  // In Vite, environment variables are accessed via import.meta.env
  // and must be prefixed with VITE_
  const value = import.meta.env[`VITE_${key}`]
  if (!value) {
    throw new Error(`Environment variable VITE_${key} is not set`)
  }
  return value
}

export const getS3Client = (): S3Client => {
  if (!s3Client) {
    const region = getEnvVar('AWS_REGION')
    const accessKeyId = getEnvVar('AWS_ACCESS_KEY_ID')
    const secretAccessKey = getEnvVar('AWS_SECRET_ACCESS_KEY')
    const endpoint = import.meta.env.VITE_AWS_ENDPOINT
    const forcePathStyle = import.meta.env.VITE_AWS_FORCE_PATH_STYLE === 'true'

    s3Client = new S3Client({
      region,
      credentials: {
        accessKeyId,
        secretAccessKey,
      },
      ...(endpoint && {
        endpoint,
        forcePathStyle,
      }),
    })
  }
  return s3Client
}
