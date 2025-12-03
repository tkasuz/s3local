import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { headObject, getObject, deleteObject, getObjectTags } from '../lib/s3Service'

export const Route = createFileRoute('/objects/$bucketName/$objectKey')({
  component: ObjectDetailPage,
})

function ObjectDetailPage() {
  const { bucketName, objectKey } = Route.useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [copySuccess, setCopySuccess] = useState(false)

  // Decode the object key (it's URL encoded in the route)
  const decodedKey = decodeURIComponent(objectKey)

  const { data: metadata, isLoading } = useQuery({
    queryKey: ['object-metadata', bucketName, decodedKey],
    queryFn: () => headObject(bucketName, decodedKey),
  })

  const { data: tags } = useQuery({
    queryKey: ['object-tags', bucketName, decodedKey],
    queryFn: () => getObjectTags(bucketName, decodedKey),
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteObject(bucketName, decodedKey),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', bucketName] })
      navigate({ to: '/buckets/$bucketName', params: { bucketName } })
    },
  })

  const handleDownload = async () => {
    try {
      const blob = await getObject(bucketName, decodedKey)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = decodedKey.split('/').pop() || decodedKey
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Download failed:', error)
      alert('Failed to download file')
    }
  }

  const handleCopyKey = async () => {
    try {
      await navigator.clipboard.writeText(decodedKey)
      setCopySuccess(true)
      setTimeout(() => setCopySuccess(false), 2000)
    } catch (error) {
      console.error('Copy failed:', error)
    }
  }

  const handleDelete = () => {
    if (confirm(`Are you sure you want to delete "${decodedKey}"?`)) {
      deleteMutation.mutate()
    }
  }

  const formatBytes = (bytes?: number) => {
    if (!bytes) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-blue-200 border-t-blue-600 mb-4"></div>
          <p className="text-gray-600 font-medium">Loading object details...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center space-x-2 text-sm">
        <Link to="/" className="text-blue-600 hover:text-blue-700 font-medium transition-colors">
          Buckets
        </Link>
        <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
        </svg>
        <Link
          to="/buckets/$bucketName"
          params={{ bucketName }}
          className="text-blue-600 hover:text-blue-700 font-medium transition-colors"
        >
          {bucketName}
        </Link>
        <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
        </svg>
        <span className="text-gray-700 font-semibold truncate max-w-md">{decodedKey.split('/').pop()}</span>
      </div>

      {/* Header */}
      <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
        <div className="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
          <div className="flex items-start space-x-4">
            <div className="flex items-center justify-center w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 shadow-lg flex-shrink-0">
              <svg className="w-7 h-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
            </div>
            <div className="flex-1 min-w-0">
              <h1 className="text-2xl font-bold text-gray-900 break-all">{decodedKey.split('/').pop()}</h1>
              <p className="text-sm text-gray-500 mt-1 break-all">{decodedKey}</p>
            </div>
          </div>

          <div className="flex flex-wrap gap-3">
            <button
              onClick={handleCopyKey}
              className="inline-flex items-center px-4 py-2.5 rounded-xl bg-white border-2 border-gray-200 hover:border-blue-300 hover:bg-blue-50 text-gray-700 font-medium transition-all shadow-sm hover:shadow"
            >
              {copySuccess ? (
                <>
                  <svg className="w-4 h-4 mr-2 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  Copied!
                </>
              ) : (
                <>
                  <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                  Copy Key
                </>
              )}
            </button>
            <button
              onClick={handleDownload}
              className="inline-flex items-center px-5 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 text-white font-medium shadow-lg shadow-blue-500/30 hover:shadow-xl hover:shadow-blue-500/40 hover:scale-105 transform transition-all"
            >
              <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              Download
            </button>
            <button
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
              className="inline-flex items-center px-4 py-2.5 rounded-xl bg-red-600 text-white font-medium shadow-lg shadow-red-500/30 hover:shadow-xl hover:shadow-red-500/40 hover:scale-105 transform transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </button>
          </div>
        </div>
      </div>

      {/* Metadata */}
      <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
        <h2 className="text-lg font-bold text-gray-900 mb-4">Object Metadata</h2>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 bg-gray-50 rounded-xl border border-gray-200">
            <p className="text-sm font-semibold text-gray-500 mb-1">Size</p>
            <p className="text-base font-medium text-gray-900">{formatBytes(metadata?.ContentLength)}</p>
          </div>
          <div className="p-4 bg-gray-50 rounded-xl border border-gray-200">
            <p className="text-sm font-semibold text-gray-500 mb-1">Content Type</p>
            <p className="text-base font-medium text-gray-900">{metadata?.ContentType || 'N/A'}</p>
          </div>
          <div className="p-4 bg-gray-50 rounded-xl border border-gray-200">
            <p className="text-sm font-semibold text-gray-500 mb-1">Last Modified</p>
            <p className="text-base font-medium text-gray-900">
              {metadata?.LastModified?.toLocaleString() || 'N/A'}
            </p>
          </div>
          <div className="p-4 bg-gray-50 rounded-xl border border-gray-200">
            <p className="text-sm font-semibold text-gray-500 mb-1">ETag</p>
            <p className="text-base font-medium text-gray-900 truncate">{metadata?.ETag || 'N/A'}</p>
          </div>
          {metadata?.StorageClass && (
            <div className="p-4 bg-gray-50 rounded-xl border border-gray-200">
              <p className="text-sm font-semibold text-gray-500 mb-1">Storage Class</p>
              <p className="text-base font-medium text-gray-900">{metadata.StorageClass}</p>
            </div>
          )}
          {metadata?.ServerSideEncryption && (
            <div className="p-4 bg-gray-50 rounded-xl border border-gray-200">
              <p className="text-sm font-semibold text-gray-500 mb-1">Encryption</p>
              <p className="text-base font-medium text-gray-900">{metadata.ServerSideEncryption}</p>
            </div>
          )}
        </div>

        {/* Custom Metadata */}
        {metadata?.Metadata && Object.keys(metadata.Metadata).length > 0 && (
          <div className="mt-6">
            <h3 className="text-base font-bold text-gray-900 mb-3">Custom Metadata</h3>
            <div className="space-y-2">
              {Object.entries(metadata.Metadata).map(([key, value]) => (
                <div key={key} className="flex items-center justify-between p-3 bg-blue-50 rounded-lg border border-blue-100">
                  <span className="text-sm font-medium text-blue-900">{key}:</span>
                  <span className="text-sm text-blue-700">{value}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Tags */}
      {tags && tags.length > 0 && (
        <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
          <h2 className="text-lg font-bold text-gray-900 mb-4">Tags</h2>
          <div className="space-y-2">
            {tags.map((tag, index) => (
              <div key={index} className="flex items-center justify-between p-3 bg-blue-50 rounded-lg border border-blue-100">
                <span className="text-sm font-medium text-blue-900">{tag.Key}:</span>
                <span className="text-sm text-blue-700">{tag.Value}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
