import { Task } from '../types';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:9090';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  };
  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (res.status === 204) return undefined as T;
  if (!res.ok) {
    const body = await res.text();
    let message: string;
    try {
      const json = JSON.parse(body);
      message = json.error || json.message || res.statusText;
    } catch {
      message = body || res.statusText;
    }
    throw new Error(message);
  }
  const text = await res.text();
  return text ? JSON.parse(text) : undefined as T;
}

export async function getTasks(): Promise<Task[]> {
  return request<Task[]>('/tasks');
}

export async function createTask(description: string): Promise<Task> {
  return request<Task>('/tasks', {
    method: 'POST',
    body: JSON.stringify({ description }),
  });
}

export async function updateTask(id: string, completed: boolean): Promise<Task> {
  return request<Task>(`/tasks/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({ completed }),
  });
}

export async function deleteTask(id: string): Promise<void> {
  await request<void>(`/tasks/${id}`, { method: 'DELETE' });
}
