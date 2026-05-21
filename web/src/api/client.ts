import axios from 'axios';

export const apiClient = axios.create({
  baseURL: '/v1',
  timeout: 15_000,
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('routex_token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});
