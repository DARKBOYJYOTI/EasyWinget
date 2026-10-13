<script>
  import { onMount, onDestroy, afterUpdate } from 'svelte';
  import { 
    Terminal, 
    X, 
    Square, 
    CheckCircle2, 
    AlertCircle,
    ArrowDown,
    Send
  } from '@lucide/svelte';
  import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime.js';
  import { CancelJob, GetJobStatus, SendJobInput, RunCommand } from '../../wailsjs/go/main/App.js';

  export let activeJob = null; // { id, title }
  export let onClose = () => {};
  export let onShowToast = (type, title, msg) => {};
  export let onJobDone = () => {};

  let terminalLines = [''];
  let renderedOutput = '';
  let isDone = true;
  let exitCode = null;
  let terminalElement;
  let elapsedSeconds = 0;
  let timerInterval;

  const defaultPrompt = 'EasyWinGet Interactive Terminal\nReady for input. Type any winget command (e.g. winget list, winget search) or shell command below:\n';

  // Smart Auto-Scroll: true by default; pauses if user scrolls up, resumes when user scrolls to bottom
  let isAutoScroll = true;
  let isScrolledUp = false;

  // Interactive input
  let interactiveInput = '';
  let inputHistory = [];
  let historyIndex = -1;

  function ansiToHtml(str) {
    if (!str) return '';
    // 1. Escape HTML
    let escaped = str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');

    // 2. Strip non-color ANSI codes (like cursor movement, clear line, etc.)
    escaped = escaped.replace(/\x1b\[[0-9;?]*[A-HJKSTfhilmnsu]/g, (match) => {
      if (match.endsWith('m')) return match;
      return '';
    });

    // 3. Convert SGR color codes (\x1b[...m)
    const colorMap = {
      '0': '</span>',
      '1': '<span style="font-weight:600">',
      '2': '<span style="opacity:0.7">',
      '30': '<span style="color:#6272a4">',
      '31': '<span style="color:#ff5555">',
      '32': '<span style="color:#50fa7b">',
      '33': '<span style="color:#f1fa8c">',
      '34': '<span style="color:#bd93f9">',
      '35': '<span style="color:#ff79c6">',
      '36': '<span style="color:#8be9fd">',
      '37': '<span style="color:#f8f8f2">',
      '90': '<span style="color:#6272a4">',
      '91': '<span style="color:#ff6e6e">',
      '92': '<span style="color:#69ff94">',
      '93': '<span style="color:#ffffa5">',
      '94': '<span style="color:#d6acff">',
      '95': '<span style="color:#ff92df">',
      '96': '<span style="color:#a4ffff">',
      '97': '<span style="color:#ffffff">',
    };

    let openSpans = 0;
    escaped = escaped.replace(/\x1b\[([0-9;]*)m/g, (_, codes) => {
      if (!codes || codes === '0') {
        const closing = '</span>'.repeat(openSpans);
        openSpans = 0;
        return closing;
      }
      let res = '';
      for (const code of codes.split(';')) {
        if (colorMap[code]) {
          if (code === '0') {
            res += '</span>'.repeat(openSpans);
            openSpans = 0;
          } else {
            res += colorMap[code];
            openSpans++;
          }
        }
      }
      return res;
    });

    escaped += '</span>'.repeat(openSpans);
    return escaped;
  }

  function resetTerminal(initialText = '') {
    terminalLines = [''];
    if (initialText) {
      appendChunk(initialText);
    } else {
      updateRendered();
    }
  }

  function appendChunk(chunk) {
    if (!chunk) return;
    let lines = [...terminalLines];
    let cur = lines[lines.length - 1];

    for (let i = 0; i < chunk.length; i++) {
      const ch = chunk[i];

      if (ch === '\r') {
        if (i + 1 < chunk.length && chunk[i + 1] === '\n') {
          // CRLF
          lines[lines.length - 1] = cur;
          lines.push('');
          cur = '';
          i++; // skip \n
        } else {
          // Standalone \r: Carriage Return! Overwrite current line for in-place CLI animations & progress bars
          lines[lines.length - 1] = cur;
          cur = '';
        }
      } else if (ch === '\n') {
        lines[lines.length - 1] = cur;
        lines.push('');
        cur = '';
      } else if (ch === '\b') {
        if (cur.length > 0) {
          cur = cur.slice(0, -1);
        }
      } else {
        cur += ch;
      }
    }

    lines[lines.length - 1] = cur;

    if (lines.length > 2500) {
      lines = lines.slice(-2000);
    }
    terminalLines = lines;
    updateRendered();
  }

  function updateRendered() {
    renderedOutput = ansiToHtml(terminalLines.join('\n'));
  }

  onMount(() => {
    // Listen for live job log streaming
    try {
      if (typeof window !== 'undefined' && window?.runtime?.EventsOn) {
        EventsOn('job:log', (data) => {
          if (activeJob && data && data.jobId === activeJob.id) {
            appendChunk(data.chunk);
          }
        });

        // Listen for job completion
        EventsOn('job:done', (data) => {
          if (activeJob && data && data.jobId === activeJob.id) {
            isDone = true;
            exitCode = data.exitCode;
            clearInterval(timerInterval);
            onJobDone();
            if (data.success) {
              onShowToast('success', 'Task Completed', `${activeJob.title} finished successfully!`);
            } else {
              onShowToast('error', 'Task Ended', `${activeJob.title} finished with code ${data.exitCode}`);
            }
          }
        });
      }
    } catch (e) {
      console.warn('TaskTerminal EventsOn error:', e);
    }

    initJob();
  });

  onDestroy(() => {
    try {
      if (typeof window !== 'undefined' && window?.runtime?.EventsOff) {
        EventsOff('job:log');
        EventsOff('job:done');
      }
    } catch (e) {}
    clearInterval(timerInterval);
  });

  afterUpdate(() => {
    if (isAutoScroll && terminalElement) {
      terminalElement.scrollTop = terminalElement.scrollHeight;
    }
  });

  function handleScroll() {
    if (!terminalElement) return;
    const { scrollTop, scrollHeight, clientHeight } = terminalElement;
    // User is at bottom if within 30px
    const atBottom = scrollHeight - scrollTop - clientHeight < 30;
    isAutoScroll = atBottom;
    isScrolledUp = !atBottom;
  }

  function scrollToBottom() {
    if (terminalElement) {
      terminalElement.scrollTop = terminalElement.scrollHeight;
      isAutoScroll = true;
      isScrolledUp = false;
    }
  }

  async function initJob() {
    if (!activeJob) {
      resetTerminal(defaultPrompt);
      isDone = true;
      return;
    }
    isDone = false;
    elapsedSeconds = 0;
    resetTerminal();
    timerInterval = setInterval(() => {
      if (!isDone) elapsedSeconds++;
    }, 1000);

    try {
      const res = await GetJobStatus(activeJob.id);
      if (res && typeof res === 'object') {
        resetTerminal(res.output || '');
        isDone = res.done;
        exitCode = res.exitCode;
        if (res.done) {
          clearInterval(timerInterval);
          onJobDone();
        }
      }
    } catch (e) {}
  }

  async function handleCancel() {
    if (!activeJob || isDone) return;
    try {
      await CancelJob(activeJob.id);
      isDone = true;
      exitCode = -1;
      clearInterval(timerInterval);
      onJobDone();
      onShowToast('warning', 'Cancellation Requested', 'Stopping task process...');
    } catch (err) {
      onShowToast('error', 'Cancel Error', String(err));
    }
  }

  async function handleSendInput() {
    const text = interactiveInput.trim();
    if (!text) return;

    interactiveInput = '';
    inputHistory = [...inputHistory, text];
    historyIndex = inputHistory.length;

    // Handle screen clear
    if (text.toLowerCase() === 'cls' || text.toLowerCase() === 'clear') {
      resetTerminal('');
      activeJob = null;
      isDone = true;
      exitCode = null;
      return;
    }

    // 1. If an active job is running, pipe stdin to it
    if (activeJob && !isDone) {
      try {
        appendChunk(`\n> ${text}\n`);
        await SendJobInput(activeJob.id, text + "\n");
        scrollToBottom();
      } catch (err) {
        appendChunk(`\n[Input Error]: ${err}\n`);
      }
      return;
    }

    // 2. Otherwise execute as a new shell/winget command
    try {
      appendChunk(`\n> ${text}\n`);
      isDone = false;
      exitCode = null;
      elapsedSeconds = 0;
      clearInterval(timerInterval);
      timerInterval = setInterval(() => {
        if (!isDone) elapsedSeconds++;
      }, 1000);

      const newJobId = await RunCommand(text);
      activeJob = { id: newJobId, title: text };
      scrollToBottom();
    } catch (err) {
      appendChunk(`\n[Command Error]: ${err}\n`);
      isDone = true;
    }
  }

  function handleKeydown(e) {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleSendInput();
    } else if (e.key === 'ArrowUp') {
      if (inputHistory.length > 0 && historyIndex > 0) {
        historyIndex--;
        interactiveInput = inputHistory[historyIndex];
      }
    } else if (e.key === 'ArrowDown') {
      if (historyIndex < inputHistory.length - 1) {
        historyIndex++;
        interactiveInput = inputHistory[historyIndex];
      } else {
        historyIndex = inputHistory.length;
        interactiveInput = '';
      }
    }
  }

  function formatTime(sec) {
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    return `${m}:${s < 10 ? '0' : ''}${s}`;
  }
</script>

<div class="terminal-view">
  <!-- Terminal Header -->
  <div class="terminal-header">
    <div class="header-left">
      <div class="terminal-icon">
        <Terminal size={17} color="var(--accent-light)" />
      </div>
      <div class="title-wrap">
        <div class="task-title">{activeJob ? activeJob.title : 'Live Console'}</div>
        <div class="task-sub">
          {#if activeJob && !isDone}
            <span class="status-running">
              <span class="status-dot"></span> Running • {formatTime(elapsedSeconds)}
            </span>
          {:else if activeJob && exitCode === 0}
            <span class="status-success">
              <CheckCircle2 size={13} /> Completed (0)
            </span>
          {:else if activeJob && exitCode !== null}
            <span class="status-error">
              <AlertCircle size={13} /> Ended ({exitCode})
            </span>
          {:else}
            <span class="status-ready">
              <span class="status-dot-ready"></span> Ready for commands
            </span>
          {/if}
        </div>
      </div>
    </div>

    <div class="header-controls">
      {#if !isDone}
        <button class="btn-danger btn-sm" on:click={handleCancel}>
          <Square size={13} /> Cancel Job
        </button>
      {/if}

      <button class="ctrl-btn" title="Close" on:click={onClose}>
        <X size={16} />
      </button>
    </div>
  </div>

  <!-- Terminal Body with Smart Auto-Scroll -->
  <div 
    class="terminal-body" 
    bind:this={terminalElement}
    on:scroll={handleScroll}
  >
    <pre class="terminal-content">{@html renderedOutput || (isDone ? defaultPrompt : 'Initializing process pipe and capturing output...')}<span class="cursor">_</span></pre>

    <!-- Floating Jump-to-Bottom Pill when user scrolls up -->
    {#if isScrolledUp}
      <button class="jump-bottom-pill" on:click={scrollToBottom}>
        <ArrowDown size={13} /> Resume Auto-scroll
      </button>
    {/if}
  </div>

  <!-- Interactive Input Console Bar -->
  <div class="interactive-bar">
    <span class="prompt-sign">></span>
    <input 
      type="text" 
      placeholder="Type command (e.g. winget list, winget search, cls) and press Enter..." 
      bind:value={interactiveInput}
      on:keydown={handleKeydown}
    />
    <button 
      class="send-btn" 
      on:click={handleSendInput} 
      disabled={!interactiveInput.trim()}
      title="Send Command / Input"
    >
      <Send size={13} />
    </button>
  </div>
</div>

<style>
  .terminal-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md);
    overflow: hidden;
    box-shadow: var(--shadow-flyout);
    position: relative;
    transition: background-color var(--transition-fast), border-color var(--transition-fast);
  }

  .terminal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 18px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border-subtle);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .terminal-icon {
    width: 32px;
    height: 32px;
    border-radius: var(--radius-sm);
    background: var(--accent-glow);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .task-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .task-sub {
    font-size: 0.75rem;
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 2px;
  }

  .status-running {
    color: var(--accent-primary);
    display: flex;
    align-items: center;
    gap: 6px;
    font-weight: 500;
  }

  .status-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--accent-primary);
    animation: pulseGlow 1s infinite alternate;
  }

  .status-success {
    color: var(--color-success);
    display: flex;
    align-items: center;
    gap: 4px;
    font-weight: 500;
  }

  .status-error {
    color: var(--color-danger);
    display: flex;
    align-items: center;
    gap: 4px;
    font-weight: 500;
  }

  .status-ready {
    color: var(--color-success-light, #22c55e);
    display: flex;
    align-items: center;
    gap: 6px;
    font-weight: 500;
  }

  .status-dot-ready {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #22c55e;
    box-shadow: 0 0 6px rgba(34, 197, 94, 0.6);
  }

  .header-controls {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .ctrl-btn {
    background: transparent;
    color: var(--text-secondary);
    border: 1px solid var(--border-subtle);
    padding: 6px;
    border-radius: var(--radius-sm);
  }
  .ctrl-btn:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.08);
  }

  .btn-sm {
    padding: 5px 12px;
    font-size: 0.8rem;
  }

  /* Terminal Body - Dark Mode default, overridden in style.css for light */
  .terminal-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;
    background: #080c14;
    position: relative;
    scroll-behavior: smooth;
  }

  .terminal-content {
    font-family: var(--font-mono);
    font-size: 0.85rem;
    line-height: 1.6;
    color: #e2e8f0;
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
  }

  .cursor {
    display: inline-block;
    color: var(--accent-primary);
    animation: blink 1s step-start infinite;
    font-weight: 700;
  }

  @keyframes blink {
    50% { opacity: 0; }
  }

  /* Floating Jump to Bottom Pill */
  .jump-bottom-pill {
    position: sticky;
    bottom: 12px;
    left: 50%;
    transform: translateX(-50%);
    margin: 0 auto;
    background: var(--bg-surface);
    border: 1px solid var(--border-card-hover);
    color: var(--accent-light);
    border-radius: var(--radius-pill);
    padding: 6px 14px;
    font-size: 0.78rem;
    font-weight: 500;
    display: flex;
    align-items: center;
    gap: 6px;
    box-shadow: var(--shadow-flyout);
    cursor: pointer;
    z-index: 10;
  }
  .jump-bottom-pill:hover {
    background: var(--bg-card-hover);
  }

  /* Interactive Command / Stdin Input Bar */
  .interactive-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 16px;
    background: var(--bg-surface);
    border-top: 1px solid var(--border-subtle);
  }

  .prompt-sign {
    color: var(--accent-primary);
    font-family: var(--font-mono);
    font-weight: 700;
    font-size: 0.95rem;
  }

  .interactive-bar input {
    flex: 1;
    background: transparent;
    border: none;
    box-shadow: none;
    padding: 4px 0;
    font-family: var(--font-mono);
    font-size: 0.85rem;
    color: var(--text-primary);
  }

  .send-btn {
    background: var(--accent-primary);
    color: var(--accent-text);
    border-radius: var(--radius-sm);
    padding: 6px 10px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .send-btn:disabled {
    opacity: 0.35;
    cursor: not-allowed;
  }
</style>
