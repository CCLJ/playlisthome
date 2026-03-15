import axios from 'axios'

const api = axios.create({
  baseURL: '/',
  withCredentials: true, // send the auth_token cookie
})

export default api
