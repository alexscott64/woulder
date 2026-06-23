import { API_BASE_URL } from './api';
import { authApiClient, authorizedFetch } from './auth';
import {
  MoneyBBox,
  MoneyFeature,
  MoneyFeatureDetail,
  MoneyFeatureFilters,
  MoneyFeatureRequest,
  MoneyNote,
  MoneyNoteRequest,
  MoneyProjectResponse,
  MoneySnapshot,
  MoneyUpload,
} from '../types/money';

function cleanFilters(filters?: MoneyFeatureFilters & { bbox?: MoneyBBox; updatedAfter?: string }) {
  const params: Record<string, string> = {};
  if (filters?.type && filters.type !== 'all') params.type = filters.type;
  if (filters?.status && filters.status !== 'all') params.status = filters.status;
  if (filters?.bbox) {
    params.bbox = `${filters.bbox.minLon},${filters.bbox.minLat},${filters.bbox.maxLon},${filters.bbox.maxLat}`;
  }
  if (filters?.updatedAfter) params.updated_after = filters.updatedAfter;
  return params;
}

export const moneyApi = {
  async getProject(slugOrId = 'money-creek'): Promise<MoneyProjectResponse> {
    const response = await authApiClient.get<MoneyProjectResponse>(`/money/projects/${slugOrId}`);
    return response.data;
  },

  async getSnapshot(projectId: string): Promise<MoneySnapshot> {
    const response = await authApiClient.get<MoneySnapshot>(`/money/projects/${projectId}/snapshot`);
    return response.data;
  },

  async listFeatures(projectId: string, filters?: MoneyFeatureFilters & { bbox?: MoneyBBox; updatedAfter?: string }): Promise<MoneyFeature[]> {
    const response = await authApiClient.get<{ features: MoneyFeature[] }>(`/money/projects/${projectId}/features`, {
      params: cleanFilters(filters),
    });
    return response.data.features;
  },

  async createFeature(projectId: string, payload: MoneyFeatureRequest): Promise<MoneyFeature> {
    const response = await authApiClient.post<MoneyFeature>(`/money/projects/${projectId}/features`, payload);
    return response.data;
  },

  async getFeature(featureId: string): Promise<MoneyFeatureDetail> {
    const response = await authApiClient.get<MoneyFeatureDetail>(`/money/features/${featureId}`);
    return response.data;
  },

  async updateFeature(featureId: string, payload: MoneyFeatureRequest): Promise<MoneyFeature> {
    const response = await authApiClient.patch<MoneyFeature>(`/money/features/${featureId}`, payload);
    return response.data;
  },

  async archiveFeature(featureId: string): Promise<void> {
    await authApiClient.delete(`/money/features/${featureId}`);
  },

  async listNotes(featureId: string): Promise<MoneyNote[]> {
    const response = await authApiClient.get<{ notes: MoneyNote[] }>(`/money/features/${featureId}/notes`);
    return response.data.notes;
  },

  async createNote(featureId: string, payload: MoneyNoteRequest): Promise<MoneyNote> {
    const response = await authApiClient.post<MoneyNote>(`/money/features/${featureId}/notes`, payload);
    return response.data;
  },

  async updateNote(noteId: string, payload: MoneyNoteRequest): Promise<MoneyNote> {
    const response = await authApiClient.patch<MoneyNote>(`/money/notes/${noteId}`, payload);
    return response.data;
  },

  async deleteNote(noteId: string): Promise<void> {
    await authApiClient.delete(`/money/notes/${noteId}`);
  },

  async uploadImage(projectId: string, file: File, options: { featureId?: string; noteId?: string } = {}): Promise<MoneyUpload> {
    const form = new FormData();
    form.append('file', file);
    if (options.featureId) form.append('feature_id', options.featureId);
    if (options.noteId) form.append('note_id', options.noteId);
    const response = await authApiClient.post<MoneyUpload>(`/money/projects/${projectId}/uploads`, form, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 60000,
    });
    return response.data;
  },

  async getUploadBlobUrl(uploadId: string): Promise<string> {
    const response = await authorizedFetch(`${API_BASE_URL}/money/uploads/${uploadId}`);
    if (!response.ok) {
      throw new Error('Failed to load image');
    }
    const blob = await response.blob();
    return URL.createObjectURL(blob);
  },

  async deleteUpload(uploadId: string): Promise<void> {
    await authApiClient.delete(`/money/uploads/${uploadId}`);
  },
};
