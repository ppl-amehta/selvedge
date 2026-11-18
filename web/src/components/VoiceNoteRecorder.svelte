<script>
  import { onMount, onDestroy } from 'svelte';

  let {
    text = $bindable(''),
    label = 'Voice Note & Transcribed Text',
    placeholder = 'Click mic to dictate voice note or type text directly...',
    rows = 3
  } = $props();

  let isListening = $state(false);
  let isSupported = $state(false);
  let statusMessage = $state('Click mic to dictate');
  let recognition = null;

  onMount(() => {
    if (typeof window !== 'undefined') {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
      if (SpeechRecognition) {
        isSupported = true;
        recognition = new SpeechRecognition();
        recognition.continuous = true;
        recognition.interimResults = true;
        recognition.lang = 'en-IN'; // Indian English / Global English

        recognition.onstart = () => {
          isListening = true;
          statusMessage = '🎙️ Listening... speak clearly';
        };

        recognition.onresult = (event) => {
          let currentTranscript = '';
          for (let i = event.resultIndex; i < event.results.length; i++) {
            const transcript = event.results[i][0].transcript;
            if (event.results[i].isFinal) {
              text = text ? text.trim() + ' ' + transcript.trim() : transcript.trim();
            } else {
              currentTranscript += transcript;
            }
          }
        };

        recognition.onerror = (event) => {
          console.warn('Speech recognition error:', event.error);
          isListening = false;
          statusMessage = `Error: ${event.error}. Tap mic to retry.`;
        };

        recognition.onend = () => {
          isListening = false;
          statusMessage = 'Dictation ended. Tap mic to resume.';
        };
      } else {
        isSupported = false;
        statusMessage = 'Speech recognition not supported in browser — type manually';
      }
    }
  });

  onDestroy(() => {
    if (recognition && isListening) {
      try {
        recognition.stop();
      } catch (e) {
        // ignore
      }
    }
  });

  function toggleListening() {
    if (!recognition) return;
    if (isListening) {
      recognition.stop();
    } else {
      try {
        recognition.start();
      } catch (err) {
        console.error('Failed to start recognition:', err);
      }
    }
  }

  function handleClear() {
    text = '';
  }
</script>

<div class="voice-recorder-container">
  <div class="header-row">
    <div class="label-area">
      <span class="field-label">{label}</span>
      <span class="status-indicator {isListening ? 'active' : ''}">
        {#if isListening}
          <span class="pulsing-dot"></span>
        {/if}
        <span class="status-text">{statusMessage}</span>
      </span>
    </div>

    <div class="btn-group">
      {#if isSupported}
        <button
          type="button"
          class="mic-btn {isListening ? 'listening' : ''}"
          onclick={toggleListening}
          title={isListening ? 'Stop listening' : 'Start voice dictation'}
          aria-label="Toggle voice dictation"
        >
          <span class="mic-icon">{isListening ? '⏹️' : '🎙️'}</span>
          <span class="mic-label">{isListening ? 'Stop Mic' : 'Voice Dictate'}</span>
        </button>
      {/if}
      {#if text}
        <button type="button" class="clear-btn" onclick={handleClear} title="Clear text">
          ✕ Clear
        </button>
      {/if}
    </div>
  </div>

  <textarea
    class="voice-textarea {isListening ? 'recording-border' : ''}"
    bind:value={text}
    {rows}
    {placeholder}
  ></textarea>
</div>

<style>
  .voice-recorder-container {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
  }

  .header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    flex-wrap: wrap;
  }

  .label-area {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .field-label {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-secondary, #94a3b8);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .status-indicator {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 0.75rem;
    color: var(--text-tertiary, #64748b);
  }

  .status-indicator.active {
    color: #ef4444;
    font-weight: 600;
  }

  .pulsing-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: #ef4444;
    animation: pulse 1s infinite alternate;
  }

  @keyframes pulse {
    0% { transform: scale(0.8); opacity: 0.6; }
    100% { transform: scale(1.3); opacity: 1; }
  }

  .btn-group {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .mic-btn {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    background: var(--bg-surface, #1e293b);
    border: 1px solid var(--border-color, #334155);
    color: var(--text-primary, #f8fafc);
    border-radius: 6px;
    padding: 4px 10px;
    font-size: 0.75rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .mic-btn:hover {
    border-color: var(--accent-color, #3b82f6);
    background: var(--bg-hover, #334155);
  }

  .mic-btn.listening {
    background: rgba(239, 68, 68, 0.15);
    border-color: #ef4444;
    color: #ef4444;
  }

  .clear-btn {
    background: transparent;
    border: none;
    color: var(--text-tertiary, #64748b);
    font-size: 0.72rem;
    cursor: pointer;
    padding: 3px 6px;
    border-radius: 4px;
  }

  .clear-btn:hover {
    color: #ef4444;
    background: rgba(239, 68, 68, 0.1);
  }

  .voice-textarea {
    width: 100%;
    padding: 10px 12px;
    font-size: 0.85rem;
    font-family: inherit;
    line-height: 1.45;
    background: var(--bg-input, #0f172a);
    color: var(--text-primary, #f8fafc);
    border: 1px solid var(--border-color, #334155);
    border-radius: 6px;
    resize: vertical;
    transition: border-color 0.2s ease, box-shadow 0.2s ease;
    box-sizing: border-box;
  }

  .voice-textarea:focus {
    outline: none;
    border-color: var(--accent-color, #3b82f6);
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
  }

  .voice-textarea.recording-border {
    border-color: #ef4444;
    box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.2);
  }
</style>
