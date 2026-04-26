import { useState, useEffect, useCallback } from 'react';
import { Task } from './types';
import { getTasks, createTask, updateTask, deleteTask } from './api/client';
import './index.css';

export default function App() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [inputError, setInputError] = useState<string | null>(null);

  const fetchTasks = useCallback(async () => {
    try {
      setError(null);
      const data = await getTasks();
      setTasks(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load tasks');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = input.trim();
    if (!trimmed) {
      setInputError('Task description cannot be empty');
      return;
    }
    setInputError(null);

    // Optimistic update
    const optimisticTask: Task = {
      id: `temp-${Date.now()}`,
      description: trimmed,
      completed: false,
    };
    setTasks(prev => [optimisticTask, ...prev]);
    setInput('');
    setSubmitting(true);

    try {
      const created = await createTask(trimmed);
      setTasks(prev =>
        prev.map(t => (t.id === optimisticTask.id ? created : t))
      );
    } catch (e) {
      // Rollback
      setTasks(prev => prev.filter(t => t.id !== optimisticTask.id));
      setInput(trimmed);
      setError(e instanceof Error ? e.message : 'Failed to create task');
    } finally {
      setSubmitting(false);
    }
  };

  const handleToggle = async (task: Task) => {
    const newCompleted = !task.completed;
    // Optimistic update
    setTasks(prev =>
      prev.map(t => (t.id === task.id ? { ...t, completed: newCompleted } : t))
    );

    try {
      const updated = await updateTask(task.id, newCompleted);
      setTasks(prev => prev.map(t => (t.id === task.id ? updated : t)));
    } catch (e) {
      // Rollback
      setTasks(prev =>
        prev.map(t => (t.id === task.id ? { ...t, completed: task.completed } : t))
      );
      setError(e instanceof Error ? e.message : 'Failed to update task');
    }
  };

  const handleDelete = async (task: Task) => {
    // Optimistic update
    setTasks(prev => prev.filter(t => t.id !== task.id));

    try {
      await deleteTask(task.id);
    } catch (e) {
      // Rollback
      setTasks(prev => {
        const exists = prev.find(t => t.id === task.id);
        if (exists) return prev;
        return [task, ...prev];
      });
      setError(e instanceof Error ? e.message : 'Failed to delete task');
    }
  };

  const activeTasks = tasks.filter(t => !t.completed);
  const completedTasks = tasks.filter(t => t.completed);

  return (
    <div className="app">
      <div className="container">
        <header className="header">
          <h1>📋 Tasks</h1>
          <p className="subtitle">Manage your daily tasks</p>
        </header>

        {error && (
          <div className="error-banner" role="alert">
            <span>{error}</span>
            <button className="dismiss-btn" onClick={() => setError(null)} aria-label="Dismiss error">×</button>
          </div>
        )}

        <form className="task-form" onSubmit={handleCreate}>
          <div className="input-group">
            <input
              type="text"
              className={`task-input${inputError ? ' input-error' : ''}`}
              placeholder="What needs to be done?"
              value={input}
              onChange={e => {
                setInput(e.target.value);
                if (inputError) setInputError(null);
              }}
              disabled={submitting}
              aria-label="New task description"
            />
            <button
              type="submit"
              className="add-btn"
              disabled={submitting}
            >
              {submitting ? 'Adding…' : 'Add Task'}
            </button>
          </div>
          {inputError && <p className="field-error">{inputError}</p>}
        </form>

        {loading ? (
          <div className="loading">Loading tasks…</div>
        ) : tasks.length === 0 ? (
          <div className="empty-state">
            <p>No tasks yet. Add one above!</p>
          </div>
        ) : (
          <div className="task-lists">
            {activeTasks.length > 0 && (
              <section className="task-section">
                <h2 className="section-title">
                  Active <span className="count">{activeTasks.length}</span>
                </h2>
                <ul className="task-list">
                  {activeTasks.map(task => (
                    <TaskRow key={task.id} task={task} onToggle={handleToggle} onDelete={handleDelete} />
                  ))}
                </ul>
              </section>
            )}

            {completedTasks.length > 0 && (
              <section className="task-section">
                <h2 className="section-title completed-title">
                  Completed <span className="count">{completedTasks.length}</span>
                </h2>
                <ul className="task-list">
                  {completedTasks.map(task => (
                    <TaskRow key={task.id} task={task} onToggle={handleToggle} onDelete={handleDelete} />
                  ))}
                </ul>
              </section>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

interface TaskRowProps {
  task: Task;
  onToggle: (task: Task) => void;
  onDelete: (task: Task) => void;
}

function TaskRow({ task, onToggle, onDelete }: TaskRowProps) {
  return (
    <li className={`task-item${task.completed ? ' completed' : ''}`}>
      <label className="task-check-label">
        <input
          type="checkbox"
          className="task-checkbox"
          checked={task.completed}
          onChange={() => onToggle(task)}
          aria-label={`Mark "${task.description}" as ${task.completed ? 'incomplete' : 'complete'}`}
        />
        <span className="checkmark" />
      </label>
      <span className={`task-description${task.completed ? ' strikethrough' : ''}`}>
        {task.description}
      </span>
      <button
        className="delete-btn"
        onClick={() => onDelete(task)}
        aria-label={`Delete task "${task.description}"`}
      >
        🗑
      </button>
    </li>
  );
}
