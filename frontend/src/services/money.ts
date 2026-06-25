import { API_BASE_URL } from './api';
import { authApiClient, authorizedFetch } from './auth';
import {
  MoneyArchiveMode,
  MoneyAreaGeometryRequest,
  MoneyAreaRequest,
  MoneyBBox,
  MoneyBoulderRequest,
  MoneyBoulderStatusRequest,
  MoneyCragSnapshot,
  MoneyFeature,
  MoneyFeatureDetail,
  MoneyFeatureFilters,
  MoneyFeatureRequest,
  MoneyMoveFeatureRequest,
  MoneyNote,
  MoneyNoteRequest,
  MoneyProblemRequest,
  MoneyProjectResponse,
  MoneySnapshot,
  MoneyTrashResponse,
  MoneyUpload,
  MoneyUploadBlockKind,
} from '../types/money';

function cleanFilters(filters?: MoneyFeatureFilters & { bbox?: MoneyBBox; updatedAfter?: string }) {
  const params: Record<string, string> = {};
  if (filters?.type && filters.type !== 'all') params.type = filters.type;
  if (filters?.status && filters.status !== 'all') params.status = filters.status;
  if (filters?.bbox) params.bbox = `${filters.bbox.minLon},${filters.bbox.minLat},${filters.bbox.maxLon},${filters.bbox.maxLat}`;
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

  async getCragSnapshot(projectId: string): Promise<MoneyCragSnapshot> {
    const response = await authApiClient.get<MoneyCragSnapshot>(`/money/projects/${projectId}/crag`);
    return response.data;
  },

  async listTrash(projectId: string): Promise<MoneyTrashResponse> {
    const response = await authApiClient.get<MoneyTrashResponse>(`/money/projects/${projectId}/trash`);
    return response.data;
  },

  async listFeatures(projectId: string, filters?: MoneyFeatureFilters & { bbox?: MoneyBBox; updatedAfter?: string }): Promise<MoneyFeature[]> {
    const response = await authApiClient.get<{ features: MoneyFeature[] }>(`/money/projects/${projectId}/features`, { params: cleanFilters(filters) });
    return response.data.features;
  },

  async createFeature(projectId: string, payload: MoneyFeatureRequest): Promise<MoneyFeature> {
    const response = await authApiClient.post<MoneyFeature>(`/money/projects/${projectId}/features`, payload);
    return response.data;
  },

  async createArea(projectId: string, payload: MoneyAreaRequest): Promise<MoneyFeature> {
    const response = await authApiClient.post<MoneyFeature>(`/money/projects/${projectId}/areas`, payload);
    return response.data;
  },

  async createBoulder(projectId: string, payload: MoneyBoulderRequest): Promise<MoneyFeature> {
    const response = await authApiClient.post<MoneyFeature>(`/money/projects/${projectId}/boulders`, payload);
    return response.data;
  },

  async createProblem(projectId: string, payload: MoneyProblemRequest): Promise<MoneyFeature> {
    const response = await authApiClient.post<MoneyFeature>(`/money/projects/${projectId}/problems`, payload);
    return response.data;
  },

  async updateBoulderStatus(featureId: string, payload: MoneyBoulderStatusRequest): Promise<MoneyFeature> {
    const response = await authApiClient.patch<MoneyFeature>(`/money/features/${featureId}/boulder-status`, payload);
    return response.data;
  },

  async updateAreaGeometry(featureId: string, payload: MoneyAreaGeometryRequest): Promise<MoneyFeature> {
    const response = await authApiClient.patch<MoneyFeature>(`/money/features/${featureId}/geometry`, payload);
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

  async archiveFeature(featureId: string, mode: MoneyArchiveMode = 'subtree'): Promise<void> {
    await authApiClient.delete(`/money/features/${featureId}`, { data: { mode } });
  },

  async moveFeatureParent(featureId: string, payload: MoneyMoveFeatureRequest): Promise<MoneyFeature> {
    const response = await authApiClient.patch<MoneyFeature>(`/money/features/${featureId}/parent`, payload);
    return response.data;
  },

  async restoreFeature(featureId: string): Promise<void> {
    await authApiClient.post(`/money/features/${featureId}/restore`);
  },

  async listProjectNotes(projectId: string): Promise<MoneyNote[]> {
    const response = await authApiClient.get<{ notes: MoneyNote[] }>(`/money/projects/${projectId}/notes`);
    return response.data.notes;
  },

  async createProjectNote(projectId: string, payload: MoneyNoteRequest): Promise<MoneyNote> {
    const response = await authApiClient.post<MoneyNote>(`/money/projects/${projectId}/notes`, payload);
    return response.data;
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

  async uploadImage(projectId: string, file: File, options: { featureId?: string; noteId?: string; blockKind?: MoneyUploadBlockKind; metadata?: Record<string, unknown> } = {}): Promise<MoneyUpload> {
    const form = new FormData();
    form.append('file', file);
    if (options.featureId) form.append('feature_id', options.featureId);
    if (options.noteId) form.append('note_id', options.noteId);
    if (options.blockKind) form.append('block_kind', options.blockKind);
    if (options.metadata) form.append('metadata', JSON.stringify(options.metadata));
    const response = await authApiClient.post<MoneyUpload>(`/money/projects/${projectId}/uploads`, form, { headers: { 'Content-Type': 'multipart/form-data' }, timeout: 60000 });
    return response.data;
  },

  async getUploadBlobUrl(uploadId: string): Promise<string> {
    const response = await authorizedFetch(`${API_BASE_URL}/money/uploads/${uploadId}`);
    if (!response.ok) throw new Error('Failed to load image');
    const blob = await response.blob();
    return URL.createObjectURL(blob);
  },

  async deleteUpload(uploadId: string): Promise<void> {
    await authApiClient.delete(`/money/uploads/${uploadId}`);
  },
};
