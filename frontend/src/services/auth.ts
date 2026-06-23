import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { API_BASE_URL } from './api';
import { AuthMeResponse, AuthResponse, MoneyCurrentUser } from '../types/money';

const REFRESH_TOKEN_KEY = 'woulder-money-refresh-token';

let accessToken: string | null = null;
let refreshPromise: Promise<AuthResponse> | null = null;

export const authApiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
});

export function getAccessToken(): string | null {
  return accessToken;
}

export function getStoredRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function setAuthTokens(tokens: Pick<AuthResponse, 'access_token' | 'refresh_token'>) {
  accessToken = tokens.access_token;
  localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token);
}

export function clearAuthTokens() {
  accessToken = null;
  localStorage.removeItem(REFRESH_TOKEN_KEY);
}

export const authApi = {
  async login(email: string, password: string): Promise<AuthResponse> {
    const response = await authApiClient.post<AuthResponse>('/auth/login', { email, password });
    setAuthTokens(response.data);
    return response.data;
  },

  async refresh(): Promise<AuthResponse> {
    const refreshToken = getStoredRefreshToken();
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }
    const response = await authApiClient.post<AuthResponse>('/auth/refresh', { refresh_token: refreshToken });
    setAuthTokens(response.data);
    return response.data;
  },

  async logout(): Promise<void> {
    const refreshToken = getStoredRefreshToken();
    try {
      if (refreshToken) {
        await authApiClient.post('/auth/logout', { refresh_token: refreshToken });
      }
    } finally {
      clearAuthTokens();
    }
  },

  async me(): Promise<MoneyCurrentUser> {
    const response = await authApiClient.get<AuthMeResponse>('/auth/me');
    return response.data.user;
  },
};

async function refreshOnce(): Promise<AuthResponse> {
  if (!refreshPromise) {
    refreshPromise = authApi.refresh().finally(() => {
      refreshPromise = null;
    });
  }
  return refreshPromise;
}

authApiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

authApiClient.interceptors.response.use(
  response => response,
  async (error: AxiosError) => {
    const original = error.config as (InternalAxiosRequestConfig & { _retry?: boolean }) | undefined;
    const isAuthEndpoint = original?.url?.includes('/auth/login') || original?.url?.includes('/auth/refresh');

    if (error.response?.status === 401 && original && !original._retry && !isAuthEndpoint && getStoredRefreshToken()) {
      original._retry = true;
      try {
        await refreshOnce();
        if (accessToken) {
          original.headers.Authorization = `Bearer ${accessToken}`;
        }
        return authApiClient(original);
      } catch (refreshError) {
        clearAuthTokens();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export async function authorizedFetch(input: RequestInfo | URL, init: RequestInit = {}): Promise<Response> {
  const headers = new Headers(init.headers);
  if (accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`);
  }

  let response = await fetch(input, { ...init, headers });
  if (response.status === 401 && getStoredRefreshToken()) {
    await refreshOnce();
    if (accessToken) {
      headers.set('Authorization', `Bearer ${accessToken}`);
    }
    response = await fetch(input, { ...init, headers });
  }
  return response;
}
