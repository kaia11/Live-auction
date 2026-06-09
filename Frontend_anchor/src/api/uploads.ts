import { apiClient, unwrapResponse } from './client'

interface UploadImageResponse {
  url: string
}

export const uploadImage = async (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  const response = await apiClient.post('/admin/uploads/image', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  })
  return unwrapResponse<UploadImageResponse>(response)
}
