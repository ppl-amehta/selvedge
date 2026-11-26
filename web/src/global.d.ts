/// <reference types="svelte" />

declare module '*.css' {
  const content: string;
  export default content;
}

declare module '*.svelte' {
  import type { Component } from 'svelte';
  const component: Component<any>;
  export default component;
}

interface Window {
  showToast?: (message: string, type?: 'success' | 'error' | 'warning' | 'info') => void;
  SpeechRecognition?: any;
  webkitSpeechRecognition?: any;
}
