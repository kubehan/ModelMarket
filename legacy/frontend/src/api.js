import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({ baseURL: '/api/v1', timeout: 30000 })

api.interceptors.request.use((cfg) => {
  const tok = localStorage.getItem('mm_token')
  if (tok) cfg.headers.Authorization = `Bearer ${tok}`
  return cfg
})

api.interceptors.response.use(
  (r) => r,
  (err) => {
    const msg = err?.response?.data?.error || err.message || '请求失败'
    if (err?.response?.status === 401) {
      localStorage.removeItem('mm_token')
      if (location.pathname.startsWith('/admin') && location.pathname !== '/admin/login') {
        location.href = '/admin/login'
      }
    }
    ElMessage.error(msg)
    return Promise.reject(err)
  }
)

export default api
