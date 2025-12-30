import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

export type TemperatureUnit = 'fahrenheit' | 'celsius';
export type SpeedUnit = 'mph' | 'kmh';

interface Settings {
  darkMode: boolean;
  temperatureUnit: TemperatureUnit;
  speedUnit: SpeedUnit;
  selectedAreaId: number | null; // null = "All Areas"
}

interface SettingsContextType {
  settings: Settings;
  updateSettings: (updates: Partial<Settings>) => void;
  toggleDarkMode: () => void;
  setSelectedArea: (areaId: number | null) => void;
}

const defaultSettings: Settings = {
  darkMode: false,
  temperatureUnit: 'fahrenheit',
  speedUnit: 'mph',
  selectedAreaId: null, // Default to "All Areas"
};

const STORAGE_KEY = 'woulder-settings';

function loadSettings(): Settings {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored);
      const settings = { ...defaultSettings, ...parsed };
      // Apply dark mode class immediately to prevent flash
      if (settings.darkMode) {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
      return settings;
    }
  } catch (e) {
    console.error('Failed to load settings from localStorage:', e);
  }
  // Ensure dark class is removed for default settings
  document.documentElement.classList.remove('dark');
  return defaultSettings;
}

function saveSettings(settings: Settings): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
  } catch (e) {
    console.error('Failed to save settings to localStorage:', e);
  }
}

const SettingsContext = createContext<SettingsContextType | undefined>(undefined);

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [settings, setSettings] = useState<Settings>(loadSettings);

  // Apply dark mode class to document
  useEffect(() => {
    if (settings.darkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [settings.darkMode]);

  // Save settings whenever they change
  useEffect(() => {
    saveSettings(settings);
  }, [settings]);

  const updateSettings = (updates: Partial<Settings>) => {
    setSettings(prev => ({ ...prev, ...updates }));
  };

  const toggleDarkMode = () => {
    setSettings(prev => {
      const newDarkMode = !prev.darkMode;
      // Immediately update the DOM class
      if (newDarkMode) {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
      return { ...prev, darkMode: newDarkMode };
    });
  };

  const setSelectedArea = (areaId: number | null) => {
    setSettings(prev => ({ ...prev, selectedAreaId: areaId }));
  };

  return (
    <SettingsContext.Provider value={{ settings, updateSettings, toggleDarkMode, setSelectedArea }}>
      {children}
    </SettingsContext.Provider>
  );
}

export function useSettings() {
  const context = useContext(SettingsContext);
  if (context === undefined) {
    throw new Error('useSettings must be used within a SettingsProvider');
  }
  return context;
}
