import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useMemo } from 'react'
import {
  listObjects,
  putObject,
  deleteObject,
  type Tag,
} from '../lib/s3Service'

export const Route = createFileRoute('/buckets/$bucketName')({
  component: BucketDetailPage,
  validateSearch: (search: Record<string, unknown>): { prefix?: string } => {
    return {
      prefix: (search.prefix as string) || '',
    }
  },
})

interface FolderItem {
  name: string
  type: 'folder' | 'file'
  fullPath: string
  size?: number
  lastModified?: Date
}

function BucketDetailPage() {
  const { bucketName } = Route.useParams()
  const navigate = useNavigate({ from: Route.fullPath })
  const { prefix = '' } = Route.useSearch()
  const queryClient = useQueryClient()
  const [showUploadModal, setShowUploadModal] = useState(false)
  const [showCreateFolderModal, setShowCreateFolderModal] = useState(false)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [objectKey, setObjectKey] = useState('')
  const [folderName, setFolderName] = useState('')
  const [tags, setTags] = useState<Tag[]>([])
  const [newTagKey, setNewTagKey] = useState('')
  const [newTagValue, setNewTagValue] = useState('')

  const { data: listResult, isLoading } = useQuery({
    queryKey: ['objects', bucketName, prefix],
    queryFn: () => listObjects(bucketName, prefix, '/'),
  })


  // Process objects to show folders and files in current prefix
  const currentItems = useMemo<FolderItem[]>(() => {
    if (!listResult) return []

    const folderMap = new Map<string, FolderItem>()
    const fileItems: FolderItem[] = []

    // Process CommonPrefixes as folders
    listResult.commonPrefixes.forEach((commonPrefix) => {
      // Extract folder name from the full prefix
      // For example: "folder1/subfolder/" -> "subfolder"
      const folderName = commonPrefix.slice(prefix.length, -1) // Remove leading prefix and trailing /
      if (folderName && !folderMap.has(folderName)) {
        folderMap.set(folderName, {
          name: folderName,
          type: 'folder',
          fullPath: commonPrefix,
        })
      }
    })

    // Process objects - check if they're folders (end with /) or files
    listResult.objects.forEach((obj) => {
      const key = obj.Key || ''

      // Skip empty keys
      if (!key) return

      // Get the relative path from current prefix
      const relativePath = key.substring(prefix.length)

      // Skip if empty (this would be the folder marker itself)
      if (!relativePath) return

      // Check if this object is a folder marker (ends with /)
      if (key.endsWith('/')) {
        // This is a folder object
        const folderName = relativePath.slice(0, -1) // Remove trailing /
        // Only add if not already in folderMap (avoid duplicates with CommonPrefixes)
        if (folderName && !folderMap.has(folderName)) {
          folderMap.set(folderName, {
            name: folderName,
            type: 'folder',
            fullPath: key,
          })
        }
      } else {
        // This is a regular file
        fileItems.push({
          name: relativePath,
          type: 'file',
          fullPath: key,
          size: obj.Size,
          lastModified: obj.LastModified,
        })
      }
    })

    // Convert folderMap to array and sort
    const folderItems = Array.from(folderMap.values()).sort((a, b) => a.name.localeCompare(b.name))

    return [
      ...folderItems,
      ...fileItems.sort((a, b) => a.name.localeCompare(b.name))
    ]
  }, [listResult, prefix])

  // Build breadcrumb path
  const breadcrumbs = useMemo(() => {
    if (!prefix) return []
    const parts = prefix.split('/').filter(Boolean)
    return parts.map((part, index) => ({
      name: part,
      path: parts.slice(0, index + 1).join('/') + '/',
    }))
  }, [prefix])

  const uploadMutation = useMutation({
    mutationFn: ({ key, file, tags }: { key: string; file: File; tags: Tag[] }) =>
      putObject(bucketName, key, file, tags),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', bucketName] })
      setShowUploadModal(false)
      setSelectedFile(null)
      setObjectKey('')
      setTags([])
    },
  })

  const createFolderMutation = useMutation({
    mutationFn: (folderPath: string) => {
      // Create an empty blob to represent the folder
      const emptyFile = new File([''], '', { type: 'application/x-directory' })
      return putObject(bucketName, folderPath, emptyFile, [])
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', bucketName] })
      setShowCreateFolderModal(false)
      setFolderName('')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (key: string) => deleteObject(bucketName, key),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['objects', bucketName] })
    },
  })


  const handleAddTag = () => {
    if (newTagKey && newTagValue) {
      setTags([...tags, { Key: newTagKey, Value: newTagValue }])
      setNewTagKey('')
      setNewTagValue('')
    }
  }

  const handleRemoveTag = (index: number) => {
    setTags(tags.filter((_, i) => i !== index))
  }

  const handleUpload = (e: React.FormEvent) => {
    e.preventDefault()
    if (selectedFile) {
      // Prepend current prefix to the object key
      const key = prefix + (objectKey || selectedFile.name)
      uploadMutation.mutate({ key, file: selectedFile, tags })
    }
  }

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setSelectedFile(file)
      if (!objectKey) {
        setObjectKey(file.name)
      }
    }
  }

  const handleFolderClick = (folderPath: string) => {
    navigate({ search: { prefix: folderPath } })
  }

  const handleBreadcrumbClick = (path: string) => {
    navigate({ search: { prefix: path } })
  }

  const handleCreateFolder = (e: React.FormEvent) => {
    e.preventDefault()
    if (folderName) {
      // Ensure folder name ends with /
      const normalizedFolderName = folderName.endsWith('/') ? folderName : folderName + '/'
      const folderPath = prefix + normalizedFolderName
      createFolderMutation.mutate(folderPath)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-blue-200 border-t-blue-600 mb-4"></div>
          <p className="text-gray-600 font-medium">Loading objects...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Back to Buckets */}
      <div className="flex items-center space-x-2 text-sm">
        <Link to="/" className="text-blue-600 hover:text-blue-700 font-medium transition-colors">
          Buckets
        </Link>
        <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
        </svg>
        <span className="text-gray-700 font-semibold">{bucketName}</span>
      </div>

      {/* Header with Breadcrumbs */}
      <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-6">
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4 mb-4">
          <div className="flex items-center space-x-4">
            <div className="flex items-center justify-center w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 shadow-lg">
              <svg className="w-7 h-7 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
            </div>
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{bucketName}</h1>
              <p className="text-sm text-gray-500 mt-0.5">
                {currentItems.length} {currentItems.length === 1 ? 'item' : 'items'}
              </p>
            </div>
          </div>

          <div className="flex flex-wrap gap-3">
            <button
              onClick={() => setShowCreateFolderModal(true)}
              className="inline-flex items-center px-4 py-2.5 rounded-xl bg-white border-2 border-gray-200 hover:border-blue-300 hover:bg-blue-50 text-gray-700 font-medium transition-all shadow-sm hover:shadow"
            >
              <svg className="w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
              </svg>
              Create Folder
            </button>
            <button
              onClick={() => setShowUploadModal(true)}
              className="inline-flex items-center px-5 py-2.5 rounded-xl bg-gradient-to-r from-blue-600 to-blue-700 text-white font-medium shadow-lg shadow-blue-500/30 hover:shadow-xl hover:shadow-blue-500/40 hover:scale-105 transform transition-all"
            >
              <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
              </svg>
              Upload
            </button>
          </div>
        </div>

        {/* Breadcrumbs */}
        <div className="flex items-center space-x-2 text-sm bg-gray-50 rounded-xl p-3 border border-gray-200">
          <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <button
            onClick={() => handleBreadcrumbClick('')}
            className={`font-medium transition-colors hover:text-blue-600 ${
              !prefix ? 'text-blue-600' : 'text-gray-600'
            }`}
          >
            Root
          </button>
          {breadcrumbs.map((crumb, index) => (
            <div key={index} className="flex items-center space-x-2">
              <svg className="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
              <button
                onClick={() => handleBreadcrumbClick(crumb.path)}
                className={`font-medium transition-colors hover:text-blue-600 ${
                  index === breadcrumbs.length - 1 ? 'text-blue-600' : 'text-gray-600'
                }`}
              >
                {crumb.name}
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Objects/Folders List */}
      {currentItems.length > 0 ? (
        <div className="bg-white rounded-2xl shadow-lg border border-gray-100 overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gradient-to-r from-gray-50 to-gray-100">
                <tr>
                  <th className="px-6 py-4 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">
                    Size
                  </th>
                  <th className="px-6 py-4 text-left text-xs font-bold text-gray-700 uppercase tracking-wider">
                    Last Modified
                  </th>
                  <th className="px-6 py-4 text-right text-xs font-bold text-gray-700 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-100">
                {currentItems.map((item) => (
                  <tr key={item.fullPath} className="hover:bg-blue-50/50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className={`flex-shrink-0 w-8 h-8 rounded-lg flex items-center justify-center mr-3 ${
                          item.type === 'folder' 
                            ? 'bg-blue-100' 
                            : 'bg-gray-100'
                        }`}>
                          {item.type === 'folder' ? (
                            <svg className="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                            </svg>
                          ) : (
                            <svg className="w-4 h-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                            </svg>
                          )}
                        </div>
                        {item.type === 'folder' ? (
                          <button
                            onClick={() => handleFolderClick(item.fullPath)}
                            className="text-sm font-medium text-blue-600 hover:text-blue-800 transition-colors"
                          >
                            {item.name}
                          </button>
                        ) : (
                          <Link
                            to="/objects/$bucketName/$objectKey"
                            params={{ bucketName, objectKey: encodeURIComponent(item.fullPath) }}
                            className="text-sm font-medium text-gray-900 hover:text-blue-600 transition-colors"
                          >
                            {item.name}
                          </Link>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600 font-medium">
                        {item.type === 'folder' ? '-' : item.size ? `${(item.size / 1024).toFixed(2)} KB` : '-'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <span className="text-sm text-gray-600">
                        {item.type === 'folder' ? '-' : item.lastModified?.toLocaleString() || '-'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right">
                      <button
                        onClick={() => {
                          const itemType = item.type === 'folder' ? 'folder' : 'file'
                          if (confirm(`Are you sure you want to delete this ${itemType} "${item.name}"?`)) {
                            deleteMutation.mutate(item.fullPath)
                          }
                        }}
                        className="inline-flex items-center px-3 py-1.5 rounded-lg text-sm font-medium text-red-600 hover:bg-red-50 transition-colors"
                      >
                        <svg className="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-2xl shadow-lg border border-gray-100 p-12 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 mb-4">
            <svg className="w-8 h-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">No items in this folder</h3>
          <p className="text-gray-600 mb-6">Upload files to get started</p>
          <button
            onClick={() => setShowUploadModal(true)}
            className="inline-flex items-center px-6 py-3 rounded-xl bg-blue-600 text-white font-medium shadow-lg hover:bg-blue-700 hover:shadow-xl transition-all"
          >
            <svg className="w-5 h-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
            Upload File
          </button>
        </div>
      )}

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20">
            <div
              className="fixed inset-0 bg-gray-900/75 backdrop-blur-sm transition-opacity"
              onClick={() => setShowUploadModal(false)}
            ></div>

            <div className="relative z-10 inline-block align-bottom bg-white rounded-2xl text-left overflow-hidden shadow-2xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full max-h-[90vh] overflow-y-auto">
              <div className="bg-gradient-to-br from-white to-gray-50 px-6 pt-6 pb-4">
                <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 shadow-md">
                      <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                      </svg>
                    </div>
                    <div>
                      <h3 className="text-xl font-bold text-gray-900">Upload File</h3>
                      {prefix && (
                        <p className="text-xs text-gray-500 mt-0.5">To: {prefix}</p>
                      )}
                    </div>
                  </div>
                  <button
                    onClick={() => setShowUploadModal(false)}
                    className="text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>

                <form onSubmit={handleUpload} className="space-y-5">
                  <div>
                    <label htmlFor="file" className="block text-sm font-semibold text-gray-700 mb-2">
                      Select File
                    </label>
                    <div className="mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-gray-300 border-dashed rounded-xl hover:border-blue-400 transition-colors">
                      <div className="space-y-1 text-center">
                        <svg className="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48">
                          <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                        <div className="flex text-sm text-gray-600">
                          <label htmlFor="file" className="relative cursor-pointer bg-white rounded-md font-medium text-blue-600 hover:text-blue-500">
                            <span>Upload a file</span>
                            <input id="file" type="file" onChange={handleFileChange} required className="sr-only" />
                          </label>
                          <p className="pl-1">or drag and drop</p>
                        </div>
                        <p className="text-xs text-gray-500">Any file type</p>
                      </div>
                    </div>
                    {selectedFile && (
                      <p className="mt-2 text-sm text-gray-600">Selected: <span className="font-medium">{selectedFile.name}</span></p>
                    )}
                  </div>

                  <div>
                    <label htmlFor="objectKey" className="block text-sm font-semibold text-gray-700 mb-2">
                      File Name
                    </label>
                    <input
                      type="text"
                      id="objectKey"
                      value={objectKey}
                      onChange={(e) => setObjectKey(e.target.value)}
                      required
                      placeholder="myfile.txt"
                      className="w-full px-4 py-3 rounded-xl border-2 border-gray-200 focus:border-blue-500 focus:ring-4 focus:ring-blue-100 transition-all outline-none"
                    />
                    {prefix && (
                      <p className="mt-1 text-xs text-gray-500">Full path: {prefix}{objectKey || selectedFile?.name || ''}</p>
                    )}
                  </div>

                  <div>
                    <label className="block text-sm font-semibold text-gray-700 mb-3">Tags (Optional)</label>
                    <div className="space-y-3">
                      {tags.map((tag, index) => (
                        <div key={index} className="flex items-center gap-2 p-3 bg-blue-50 rounded-lg border border-blue-100">
                          <div className="flex-1 flex items-center gap-2 text-sm">
                            <span className="font-medium text-blue-900">{tag.Key}:</span>
                            <span className="text-blue-700">{tag.Value}</span>
                          </div>
                          <button
                            type="button"
                            onClick={() => handleRemoveTag(index)}
                            className="text-red-500 hover:text-red-700 hover:bg-red-100 p-1 rounded transition-colors"
                          >
                            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                          </button>
                        </div>
                      ))}
                      <div className="flex gap-2">
                        <input
                          type="text"
                          placeholder="Key"
                          value={newTagKey}
                          onChange={(e) => setNewTagKey(e.target.value)}
                          className="flex-1 px-3 py-2 rounded-lg border-2 border-gray-200 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 transition-all outline-none text-sm"
                        />
                        <input
                          type="text"
                          placeholder="Value"
                          value={newTagValue}
                          onChange={(e) => setNewTagValue(e.target.value)}
                          className="flex-1 px-3 py-2 rounded-lg border-2 border-gray-200 focus:border-blue-500 focus:ring-2 focus:ring-blue-100 transition-all outline-none text-sm"
                        />
                        <button
                          type="button"
                          onClick={handleAddTag}
                          className="px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-lg font-medium text-sm text-gray-700 transition-colors"
                        >
                          Add
                        </button>
                      </div>
                    </div>
                  </div>

                  <div className="flex gap-3 pt-4">
                    <button
                      type="submit"
                      disabled={uploadMutation.isPending}
                      className="flex-1 px-5 py-3 bg-gradient-to-r from-blue-600 to-blue-700 text-white font-semibold rounded-xl shadow-lg hover:shadow-xl hover:scale-105 transform transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
                    >
                      {uploadMutation.isPending ? 'Uploading...' : 'Upload'}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setShowUploadModal(false)
                        setSelectedFile(null)
                        setObjectKey('')
                        setTags([])
                      }}
                      className="px-5 py-3 bg-gray-100 hover:bg-gray-200 text-gray-700 font-semibold rounded-xl transition-colors"
                    >
                      Cancel
                    </button>
                  </div>

                  {uploadMutation.isError && (
                    <div className="rounded-xl bg-red-50 border border-red-200 p-4">
                      <p className="text-sm text-red-700">{(uploadMutation.error as Error).message}</p>
                    </div>
                  )}
                </form>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Create Folder Modal */}
      {showCreateFolderModal && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20">
            <div
              className="fixed inset-0 bg-gray-900/75 backdrop-blur-sm transition-opacity"
              onClick={() => setShowCreateFolderModal(false)}
            ></div>

            <div className="relative z-10 inline-block align-bottom bg-white rounded-2xl text-left overflow-hidden shadow-2xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <div className="bg-gradient-to-br from-white to-gray-50 px-6 pt-6 pb-4">
                <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-blue-600 shadow-md">
                      <svg className="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
                      </svg>
                    </div>
                    <div>
                      <h3 className="text-xl font-bold text-gray-900">Create Folder</h3>
                      {prefix && (
                        <p className="text-xs text-gray-500 mt-0.5">In: {prefix}</p>
                      )}
                    </div>
                  </div>
                  <button
                    onClick={() => setShowCreateFolderModal(false)}
                    className="text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>

                <form onSubmit={handleCreateFolder} className="space-y-5">
                  <div>
                    <label htmlFor="folderName" className="block text-sm font-semibold text-gray-700 mb-2">
                      Folder Name
                    </label>
                    <input
                      type="text"
                      id="folderName"
                      value={folderName}
                      onChange={(e) => setFolderName(e.target.value)}
                      required
                      placeholder="my-folder"
                      className="w-full px-4 py-3 rounded-xl border-2 border-gray-200 focus:border-blue-500 focus:ring-4 focus:ring-blue-100 transition-all outline-none"
                    />
                    {folderName && (
                      <p className="mt-1 text-xs text-gray-500">
                        Full path: {prefix}{folderName}{folderName.endsWith('/') ? '' : '/'}
                      </p>
                    )}
                  </div>

                  <div className="flex gap-3 pt-4">
                    <button
                      type="submit"
                      disabled={createFolderMutation.isPending}
                      className="flex-1 px-5 py-3 bg-gradient-to-r from-blue-600 to-blue-700 text-white font-semibold rounded-xl shadow-lg hover:shadow-xl hover:scale-105 transform transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
                    >
                      {createFolderMutation.isPending ? 'Creating...' : 'Create Folder'}
                    </button>
                    <button
                      type="button"
                      onClick={() => {
                        setShowCreateFolderModal(false)
                        setFolderName('')
                      }}
                      className="px-5 py-3 bg-gray-100 hover:bg-gray-200 text-gray-700 font-semibold rounded-xl transition-colors"
                    >
                      Cancel
                    </button>
                  </div>

                  {createFolderMutation.isError && (
                    <div className="rounded-xl bg-red-50 border border-red-200 p-4">
                      <p className="text-sm text-red-700">{(createFolderMutation.error as Error).message}</p>
                    </div>
                  )}
                </form>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
