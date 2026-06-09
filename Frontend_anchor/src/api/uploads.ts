import { apiClient, unwrapResponse, type ApiEnvelope } from './client'
import { AxiosError } from 'axios'

interface UploadImageResponse {
  url: string
}

export const uploadImage = async (file: File) => {
  const formData = new FormData()
  formData.append('file', file)
  try {
    const response = await apiClient.post('/admin/uploads/image', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return unwrapResponse<UploadImageResponse>(response)
  } catch (error) {
    const err = error as AxiosError<ApiEnvelope<null>>
    const apiMessage = err.response?.data?.message
    if (apiMessage) {
      throw new Error(apiMessage)
    }
    throw error
  }
}
