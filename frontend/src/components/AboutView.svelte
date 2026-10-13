<script>
  import { 
    Info, 
    Heart, 
    ShieldCheck, 
    Scale, 
    BookOpen, 
    Cpu, 
    ExternalLink, 
    Copy, 
    Check, 
    AlertTriangle, 
    Sparkles, 
    Package, 
    Terminal, 
    HardDrive, 
    Download, 
    RefreshCw,
    FolderCheck,
    Layers,
    Search,
    SlidersHorizontal
  } from '@lucide/svelte';
  import { onMount } from 'svelte';
  import { BrowserOpenURL } from '../../wailsjs/runtime/runtime.js';
  import { OpenURLInDefaultBrowser, IsPortableMode } from '../../wailsjs/go/main/App.js';

  export let version = '4.0.0';
  export let onShowToast = (type, title, msg) => {};

  let activeSection = 'guide'; // 'guide' | 'tos' | 'privacy' | 'diagnostics'
  let copiedDiagnostics = false;
  let isPortable = false;

  onMount(async () => {
    try {
      isPortable = await IsPortableMode();
    } catch (e) {}
  });

  const diagnosticsText = `Application: EasyWinGet Desktop
Version: v${version} Native Desktop
Storage Engine: SQLite 3 (CGO-free Pure Go WAL Mode)
Core Architecture: Go 1.25+ Wails v2 + Svelte 5
Package Manager Engine: Windows Package Manager CLI (winget)
Execution Architecture: Pure Go Native Execution (Zero PowerShell Script Dependency)
Registry Engine: Native Win32 Registry Scanner (BCUninstaller Style)
Platform: Windows 11 / 10 64-bit (Win32 DWM Mica Translucency)
Developer: Jyoti Karmakar
Repository: https://github.com/DARKBOYJYOTI/EasyWinget`;

  function copyDiagnostics() {
    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText(diagnosticsText);
      copiedDiagnostics = true;
      onShowToast('success', 'Copied to Clipboard', 'System runtime diagnostics copied');
      setTimeout(() => { copiedDiagnostics = false; }, 2500);
    }
  }

  async function openExternal(url) {
    if (!url) return;
    try {
      await OpenURLInDefaultBrowser(url);
    } catch (e) {
      try {
        BrowserOpenURL(url);
      } catch (err) {
        if (typeof window !== 'undefined') window.open(url, '_blank');
      }
    }
  }
</script>

<div class="about-page">
  <div class="about-container">
    
    <!-- Hero Branding Card -->
    <div class="hero-card">
      <div class="hero-glow"></div>
      
      <div class="hero-header-row">
        <div class="hero-logo-box">
          <Package size={36} color="var(--accent-light, #60a5fa)" />
        </div>
        <div class="hero-title-group">
          <div class="title-with-badge">
            <h1 class="app-brand-title">EasyWinGet</h1>
            <span class="version-badge">v{version} {isPortable ? 'Portable Edition' : 'Native Desktop'}</span>
          </div>
          <p class="app-tagline">Next-Generation Native Windows Package Management Hub</p>
        </div>
      </div>

      <p class="hero-description">
        Engineered with a high-performance pure Go backend and lightweight Svelte 5 frontend, EasyWinGet delivers sub-millisecond local SQLite caching, real-time goroutine process streaming, native Windows registry package discovery, and seamless offline installer workflows.
      </p>

      <div class="hero-actions">
        <button class="action-btn primary" on:click={() => openExternal('https://github.com/DARKBOYJYOTI/EasyWinget')}>
          <ExternalLink size={15} />
          <span>GitHub Repository</span>
        </button>
        <button class="action-btn secondary" on:click={() => openExternal('https://github.com/DARKBOYJYOTI/EasyWinget/issues')}>
          <AlertTriangle size={15} />
          <span>Report a Bug</span>
        </button>
        <button class="action-btn ghost" on:click={copyDiagnostics}>
          {#if copiedDiagnostics}
            <Check size={15} color="#10b981" />
            <span style="color: #10b981;">Copied!</span>
          {:else}
            <Copy size={15} />
            <span>Copy Diagnostics</span>
          {/if}
        </button>
      </div>
    </div>

    <!-- Technical Architecture Specifications Card -->
    <div class="diagnostics-card">
      <div class="diag-header">
        <div class="diag-icon-wrapper">
          <Cpu size={19} color="currentColor" />
        </div>
        <div>
          <h2 class="diag-title">System Runtime & Architecture Diagnostics</h2>
          <p class="diag-subtitle">Core engine specifications and native Windows platform integration</p>
        </div>
      </div>

      <div class="specs-grid">
        <div class="spec-card">
          <div class="spec-label">Application Version</div>
          <div class="spec-value highlight">v{version} Native Desktop</div>
          <div class="spec-desc">Zero-dependency standalone Windows 64-bit binary</div>
        </div>

        <div class="spec-card">
          <div class="spec-label">Local Storage Engine</div>
          <div class="spec-value">SQLite 3 (WAL Mode)</div>
          <div class="spec-desc">CGO-free pure Go modern SQLite with sub-millisecond query latency</div>
        </div>

        <div class="spec-card">
          <div class="spec-label">Core Architecture</div>
          <div class="spec-value">Go 1.25+ Wails v2 + Svelte 5</div>
          <div class="spec-desc">Zero external browser runtime overhead with native Win32 Webview2 bindings</div>
        </div>

        <div class="spec-card">
          <div class="spec-label">Package Manager Engine</div>
          <div class="spec-value">Windows Package Manager (winget)</div>
          <div class="spec-desc">Official Microsoft WinGet CLI automation with UTF-8 stream processing</div>
        </div>

        <div class="spec-card">
          <div class="spec-label">Registry Engine</div>
          <div class="spec-value">BCUninstaller-Style Scanner</div>
          <div class="spec-desc">Zero-PowerShell native 64/32-bit Windows Registry inspection and uninstaller</div>
        </div>

        <div class="spec-card">
          <div class="spec-label">Process Subsystem</div>
          <div class="spec-value">Goroutine Live Log Streamer</div>
          <div class="spec-desc">Real-time Wails event bus broadcasting stdout/stderr with cancellation</div>
        </div>
      </div>
    </div>

    <!-- Active Development & Bug Reporting Notice -->
    <div class="dev-notice-card">
      <div class="notice-icon">
        <AlertTriangle size={24} color="#f59e0b" />
      </div>
      <div class="notice-content">
        <h3 class="notice-heading">Under Active Development — Help Us Improve!</h3>
        <p class="notice-text">
          EasyWinGet is continuously evolving with new features, optimizations, and UI refinements. While extensively tested, occasional bugs, display anomalies, or platform edge cases may occur. If you spot any unexpected behavior or have ideas for enhancements, please open an issue on our GitHub tracker. Every bug report directly shapes future releases!
        </p>
      </div>
      <button class="notice-btn" on:click={() => openExternal('https://github.com/DARKBOYJYOTI/EasyWinget/issues')}>
        <span>Create Issue</span>
        <ExternalLink size={14} />
      </button>
    </div>

    <!-- Section Navigation Tabs -->
    <div class="section-tabs-bar">
      <button 
        class="tab-pill {activeSection === 'guide' ? 'active' : ''}"
        on:click={() => activeSection = 'guide'}
      >
        <BookOpen size={16} />
        <span>Help & User Guide</span>
      </button>

      <button 
        class="tab-pill {activeSection === 'tos' ? 'active' : ''}"
        on:click={() => activeSection = 'tos'}
      >
        <Scale size={16} />
        <span>Terms of Service</span>
      </button>

      <button 
        class="tab-pill {activeSection === 'privacy' ? 'active' : ''}"
        on:click={() => activeSection = 'privacy'}
      >
        <ShieldCheck size={16} />
        <span>Privacy Policy</span>
      </button>
    </div>

    <!-- Tab Content Container -->
    <div class="tab-content-panel">

      <!-- 1. HELP & USER GUIDE -->
      {#if activeSection === 'guide'}
        <div class="guide-section">
          <div class="section-intro">
            <h2>User Guide & Comprehensive Help</h2>
            <p>Master package installation, updates management, offline installer archiving, and system troubleshooting.</p>
          </div>

          <div class="guide-grid">
            <div class="guide-card">
              <div class="guide-card-header">
                <Search size={20} class="guide-icon" />
                <h3>1. Discover & Install Packages</h3>
              </div>
              <p>
                Browse through more than 10,000 community packages curated from the official Windows Package Manager and Microsoft Store repositories.
              </p>
              <ul class="guide-list">
                <li><strong>Search Bar:</strong> Type application names, IDs, publishers, or keywords.</li>
                <li><strong>Category Pills:</strong> Quickly filter by Development, Gaming, Utilities, Media, and Productivity.</li>
                <li><strong>Install Button:</strong> Initiates silent unattended setup. If administrative elevation is required, Windows UAC will seamlessly prompt.</li>
                <li><strong>Details Modal:</strong> Click any package card to view license, publisher URL, installer SHA256 hashes, and tags.</li>
              </ul>
            </div>

            <div class="guide-card">
              <div class="guide-card-header">
                <Package size={20} class="guide-icon" />
                <h3>2. Manage Installed Applications</h3>
              </div>
              <p>
                The Installed tab aggregates every Win32 desktop application and UWP modern Windows app installed on your computer.
              </p>
              <ul class="guide-list">
                <li><strong>Real-time Filtering:</strong> Search installed programs instantly with zero lag.</li>
                <li><strong>App Information:</strong> Inspect exact install dates, registry uninstall keys, and publishers.</li>
                <li><strong>Native Uninstallation:</strong> Trigger clean uninstallers directly from the app card or details modal.</li>
                <li><strong>JSON Backup Export:</strong> In Settings, export all your installed software to a single JSON file to quickly migrate setups to new PCs.</li>
              </ul>
            </div>

            <div class="guide-card">
              <div class="guide-card-header">
                <RefreshCw size={20} class="guide-icon" />
                <h3>3. Smart Software Updates</h3>
              </div>
              <p>
                Keep all your desktop software updated securely with automated version comparison and differential detection.
              </p>
              <ul class="guide-list">
                <li><strong>Update All:</strong> Update all outdated programs sequentially with a single click.</li>
                <li><strong>Ignore Updates:</strong> Want to keep a specific version pinned? Click the menu icon and choose "Ignore Update" to exclude it from future scans.</li>
                <li><strong>Ignored Hub:</strong> View and unignore packages anytime from the dedicated tab at the top of Updates view.</li>
              </ul>
            </div>

            <div class="guide-card">
              <div class="guide-card-header">
                <Download size={20} class="guide-icon" />
                <h3>4. Offline Installer Downloads</h3>
              </div>
              <p>
                Download standalone installer files (<code>.exe</code>, <code>.msi</code>) for offline storage, archiving, or air-gapped deployment.
              </p>
              <ul class="guide-list">
                <li><strong>Storage Directory:</strong> Saved in <code>%USERPROFILE%\Downloads\EasyWinGet</code>.</li>
                <li><strong>Open Folder:</strong> Directly highlights the specific downloaded file inside Windows File Explorer.</li>
                <li><strong>Offline Install:</strong> Run downloaded setups offline anytime with elevated execution.</li>
              </ul>
            </div>

            <div class="guide-card">
              <div class="guide-card-header">
                <Terminal size={20} class="guide-icon" />
                <h3>5. Live Console & Interactive Shell</h3>
              </div>
              <p>
                Full visibility into Windows Package Manager output, stdout/stderr streams, and interactive command execution.
              </p>
              <ul class="guide-list">
                <li><strong>Smart Auto-Scroll:</strong> Follows logs live as they stream from background workers.</li>
                <li><strong>Interactive Input:</strong> Type custom inputs or responses (e.g. <code>Y</code> / <code>N</code>) directly into running winget jobs.</li>
                <li><strong>Custom Commands:</strong> Execute custom winget or powershell parameters straight from the terminal.</li>
                <li><strong>Job Cancellation:</strong> Gracefully terminate any active task at any moment without freezing the UI.</li>
              </ul>
            </div>

            <div class="guide-card">
              <div class="guide-card-header">
                <SlidersHorizontal size={20} class="guide-icon" />
                <h3>6. Windows Integration & Autostart</h3>
              </div>
              <p>
                Customize application startup, theme colorization, and system tray background running.
              </p>
              <ul class="guide-list">
                <li><strong>Silent Boot:</strong> Automatically start EasyWinGet minimized in the Windows System Tray on login.</li>
                <li><strong>Minimize on Close:</strong> Keeps EasyWinGet accessible in the Action Center tray when closing the window.</li>
                <li><strong>Windows 11 Accent:</strong> Seamlessly sync the app's primary color with your Windows Personalization accent.</li>
                <li><strong>Wipe SQLite Cache:</strong> Instantly reset cached tables to pull completely fresh data from winget.</li>
              </ul>
            </div>
          </div>
        </div>

      <!-- 2. TERMS OF SERVICE (INDIAN LAW FULL LAWYER MODE) -->
      {:else if activeSection === 'tos'}
        <div class="legal-document">
          <div class="legal-header">
            <h2>Terms of Service</h2>
            <div class="legal-statutory-badge">Governed by the Laws of the Republic of India</div>
            <p class="legal-date">Last Updated: September 2026 • Effective Immediately</p>
          </div>

          <div class="legal-body">
            <div class="legal-clause">
              <h4>1. STATUTORY FRAMEWORK & ELECTRONIC RECORD</h4>
              <p>
                This document is an electronic record generated in terms of the <strong>Information Technology Act, 2000 (Act No. 21 of 2000)</strong>, the <strong>Indian Contract Act, 1872 (Act No. 9 of 1872)</strong>, and the rules framed thereunder as applicable. This electronic record is generated by a computer system and does not require any physical or digital signatures in accordance with Section 10A of the Information Technology Act, 2000.
              </p>
              <p>
                By downloading, installing, launching, executing, or using <strong>EasyWinGet</strong> ("the Application"), you ("the User") unequivocally agree to be bound by these Terms of Service. If you do not agree to these terms, you must immediately cease use of and uninstall the Application.
              </p>
            </div>

            <div class="legal-clause">
              <h4>2. NATURE OF THE APPLICATION & LICENSE GRANT</h4>
              <p>
                EasyWinGet is developed and authored by <strong>Jyoti Karmakar</strong> ("the Author" / "Developer") as a personal, non-commercial, open-source utility software. The Author hereby grants the User a revocable, non-exclusive, non-transferable, royalty-free, and limited license to install and execute the Application solely for lawful software management purposes on Windows-based personal computer systems.
              </p>
            </div>

            <div class="legal-clause">
              <h4>3. INTERMEDIARY SAFE HARBOR & THIRD-PARTY PACKAGES (SECTION 79, IT ACT 2000)</h4>
              <p>
                The User expressly acknowledges that EasyWinGet functions strictly as a client-side graphical orchestration interface (GUI) and local wrapper for Microsoft Corporation's command-line tool <code>winget.exe</code> (Windows Package Manager).
              </p>
              <p>
                <strong>(a) No Hosting or Endorsement:</strong> EasyWinGet does not host, upload, compile, verify, maintain, or distribute third-party application binaries, installer executables, or repository catalogs. All package metadata, repositories, and installation scripts originate from Microsoft Corporation or respective third-party publishers.
              </p>
              <p>
                <strong>(b) Intermediary Exemption:</strong> In accordance with <strong>Section 79 of the Information Technology Act, 2000</strong> and the <em>Information Technology (Intermediary Guidelines and Digital Media Ethics Code) Rules, 2021</em>, the Author and EasyWinGet claim full statutory intermediary immunity and safe harbor. Under no circumstances shall the Author be liable for the availability, authenticity, licensing terms, copyright violations, malware, vulnerabilities, or software defects of any third-party program installed or managed via EasyWinGet.
              </p>
              <p>
                <strong>(c) Third-Party EULAs:</strong> Each application installed via EasyWinGet is governed strictly by the End User License Agreement (EULA) and terms established by its respective publisher. The User remains solely responsible for reviewing and accepting individual package agreements.
              </p>
            </div>

            <div class="legal-clause">
              <h4>4. COMPREHENSIVE DISCLAIMER OF WARRANTIES ("AS IS" & "AS AVAILABLE")</h4>
              <p>
                TO THE MAXIMUM EXTENT PERMITTED UNDER SECTION 16 AND SECTION 17 OF THE SALE OF GOODS ACT, 1930 AND APPLICABLE INDIAN JURISPRUDENCE, THE APPLICATION IS PROVIDED TO THE USER STRICTLY ON AN <strong>"AS IS"</strong> AND <strong>"AS AVAILABLE"</strong> BASIS, WITH ALL FAULTS AND DEFECTS, WITHOUT WARRANTY OF ANY KIND.
              </p>
              <p>
                THE AUTHOR EXPRESSLY DISCLAIMS ALL WARRANTIES, WHETHER STATUTORY, EXPRESS, IMPLIED, OR COLLATERAL, INCLUDING BUT NOT LIMITED TO IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, OPERATIONAL CONTINUITY, SYSTEM COMPATIBILITY, ERROR-FREE EXECUTION, FREEDOM FROM PROGRAMMING BUGS, REGISTRY INTEGRITY, OR NON-INFRINGEMENT OF THIRD-PARTY INTELLECTUAL PROPERTY RIGHTS.
              </p>
            </div>

            <div class="legal-clause">
              <h4>5. LIMITATION OF LIABILITY</h4>
              <p>
                IN NO EVENT SHALL JYOTI KARMAKAR, CONTRIBUTORS, OR AFFILIATED THIRD PARTIES BE LIABLE UNDER CONTRACT, TORT (INCLUDING NEGLIGENCE), STRICT LIABILITY, STATUTORY DUTY, OR ANY OTHER LEGAL OR EQUITABLE DOCTRINE FOR ANY DIRECT, INDIRECT, INCIDENTAL, CONSEQUENTIAL, SPECIAL, PUNITIVE, OR EXEMPLARY DAMAGES.
              </p>
              <p>
                THIS LIMITATION ENCOMPASSES, WITHOUT LIMITATION, DAMAGES FOR LOSS OF DATA, CORRUPTION OF OPERATING SYSTEM FILES OR WINDOWS REGISTRY ENTRIES, HARDWARE MALFUNCTION, SYSTEM CRASHES, LOSS OF BUSINESS REVENUE OR PROFITS, BUSINESS INTERRUPTION, OR UNFORESEEN DOWNTIME ARISING OUT OF OR IN CONNECTION WITH THE USE OF, MISUSE OF, OR INABILITY TO USE THE APPLICATION.
              </p>
            </div>

            <div class="legal-clause">
              <h4>6. GOVERNING LAW, ARBITRATION & EXCLUSIVE JURISDICTION</h4>
              <p>
                These Terms of Service and any contractual or non-contractual claims arising out of or related to EasyWinGet shall be construed, interpreted, and governed exclusively in accordance with the substantive laws of the <strong>Republic of India</strong>, without regard to conflict of law principles.
              </p>
              <p>
                Any dispute, difference, controversy, or claim arising out of or relating to this agreement shall be subject to the exclusive jurisdiction of the competent civil courts located in <strong>West Bengal, India</strong>, and the User hereby consents and submits to the personal jurisdiction of such courts.
              </p>
            </div>
          </div>
        </div>

      <!-- 3. PRIVACY POLICY (INDIAN LAW DPDP ACT 2023 & SPDI RULES 2011) -->
      {:else if activeSection === 'privacy'}
        <div class="legal-document">
          <div class="legal-header">
            <h2>Privacy Policy</h2>
            <div class="legal-statutory-badge">Compliant with the Digital Personal Data Protection Act, 2023 (DPDP Act)</div>
            <p class="legal-date">Last Updated: September 2026 • 100% Zero-Telemetry Policy</p>
          </div>

          <div class="legal-body">
            <div class="legal-clause">
              <h4>1. LEGAL FRAMEWORK & COMMITMENT TO PRIVACY</h4>
              <p>
                This Privacy Policy is formulated in strict compliance with the <strong>Digital Personal Data Protection Act, 2023 (Act No. 22 of 2023)</strong> and Rule 3 & 4 of the <strong>Information Technology (Reasonable Security Practices and Procedures and Sensitive Personal Data or Information) Rules, 2011 (SPDI Rules)</strong> framed under the Information Technology Act, 2000.
              </p>
              <p>
                The Developer, <strong>Jyoti Karmakar</strong>, places absolute priority on user privacy, software sovereignty, and data autonomy. EasyWinGet operates entirely on an offline-first, client-side, zero-telemetry architecture.
              </p>
            </div>

            <div class="legal-clause">
              <h4>2. ZERO PERSONAL DATA HARVESTING (DATA MINIMIZATION)</h4>
              <p>
                In strict accordance with the principle of <em>Data Minimization</em> under Section 6 of the DPDP Act, 2023, <strong>EasyWinGet DOES NOT COLLECT, HARVEST, LOG, TRANSMIT, PROFILE, OR SELL ANY PERSONAL DATA OR IDENTIFIABLE INFORMATION</strong>.
              </p>
              <ul class="legal-list">
                <li><strong>No Identifiers:</strong> We do not collect names, email addresses, phone numbers, IP addresses, MAC addresses, device IDs, or hardware GUIDs.</li>
                <li><strong>No Telemetry:</strong> There are zero analytics SDKs, tracking pixels, crash beacons, or user monitoring scripts embedded in the application.</li>
                <li><strong>No Advertising:</strong> EasyWinGet is completely free of third-party advertisements, tracking cookies, and commercial monetization.</li>
              </ul>
            </div>

            <div class="legal-clause">
              <h4>3. 100% LOCAL-FIRST CLIENT STORAGE</h4>
              <p>
                All data necessary for the operation of EasyWinGet is processed and stored strictly on the User's local machine within an embedded CGO-free pure Go SQLite database (<code>easywinget.db</code>).
              </p>
              <ul class="legal-list">
                <li><strong>Cached Catalog:</strong> Search queries and metadata fetched from Microsoft WinGet are cached locally to provide sub-millisecond search performance.</li>
                <li><strong>User Preferences:</strong> Stored locally, including theme settings (Dark, Light, Auto), Windows accent preferences, startup choices, and ignored update IDs.</li>
                <li><strong>Complete User Sovereignty:</strong> You retain absolute control over your local data. At any time, you can purge the entirety of the local SQLite database via Settings &rarr; <em>"Wipe SQLite Local Cache"</em>.</li>
              </ul>
            </div>

            <div class="legal-clause">
              <h4>4. DISCLOSURE OF EXTERNAL NETWORK TRANSMISSIONS</h4>
              <p>
                To provide its core package management functionality, EasyWinGet communicates solely with the following publicly available, legitimate web endpoints:
              </p>
              <ul class="legal-list">
                <li><strong>Microsoft WinGet CDN:</strong> Connects to Microsoft's official content delivery network (<code>cdn.winget.microsoft.com</code>) to download package manifests, catalog updates, and installer binaries initiated directly by the User.</li>
                <li><strong>Public Favicon Resolvers:</strong> High-resolution application brand icons are fetched over encrypted HTTPS from the Google S2 Favicon API and DuckDuckGo Favicon Service using public domain names (e.g. <code>google.com</code>, <code>mozilla.org</code>). No user data or system info is included in these requests.</li>
                <li><strong>Connectivity Health Checks:</strong> Periodic lightweight HTTP HEAD requests are made to standard endpoints (<code>msftconnecttest.com</code> / <code>1.1.1.1</code>) to detect whether active internet access is present.</li>
              </ul>
            </div>

            <div class="legal-clause">
              <h4>5. STATUTORY GRIEVANCE REDRESSAL OFFICER</h4>
              <p>
                In accordance with Rule 5(9) of the Information Technology (Reasonable Security Practices and Procedures and Sensitive Personal Data or Information) Rules, 2011 and Section 12 of the Digital Personal Data Protection Act, 2023, the details of the designated Grievance Officer are as follows:
              </p>
              <div class="grievance-card">
                <div class="grievance-title">Grievance & Privacy Officer</div>
                <div class="grievance-name">Jyoti Karmakar</div>
                <div class="grievance-role">Author, Creator & Maintainer — EasyWinGet</div>
                <div class="grievance-contact">Contact & Issues: Official GitHub Repository (<a href="https://github.com/DARKBOYJYOTI/EasyWinget/issues" on:click|preventDefault={() => openExternal('https://github.com/DARKBOYJYOTI/EasyWinget/issues')}>https://github.com/DARKBOYJYOTI/EasyWinget/issues</a>)</div>
                <div class="grievance-jurisdiction">Jurisdiction: West Bengal, Republic of India</div>
              </div>
            </div>
          </div>
        </div>
      {/if}

    </div>

    <!-- The Iconic Requested Blinking Heart Footer -->
    <div class="about-footer-signature">
      <p class="signature-text">
        Made with <span class="blinking-heart">❤️</span> by <span class="author-name">Jyoti Karmakar</span>
      </p>
      <div class="signature-subtext">EasyWinGet Desktop • Free & Open-Source Software</div>
    </div>

  </div>
</div>

<style>
  .about-page {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 100%;
    height: 100%;
    overflow-y: auto;
    padding: 32px 24px 72px;
    user-select: text;
  }

  .about-container {
    width: 100%;
    max-width: 820px;
    display: flex;
    flex-direction: column;
    gap: 22px;
  }

  /* Hero Card */
  .hero-card {
    position: relative;
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg, 12px);
    padding: 28px 30px;
    overflow: hidden;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  }

  .hero-glow {
    position: absolute;
    top: -40px;
    right: -40px;
    width: 180px;
    height: 180px;
    background: radial-gradient(circle, var(--accent-light, #0078d4) 0%, transparent 70%);
    opacity: 0.18;
    pointer-events: none;
    filter: blur(35px);
  }

  .hero-header-row {
    display: flex;
    align-items: center;
    gap: 18px;
    margin-bottom: 14px;
  }

  .hero-logo-box {
    width: 60px;
    height: 60px;
    border-radius: 14px;
    background: linear-gradient(135deg, rgba(255, 255, 255, 0.08) 0%, rgba(255, 255, 255, 0.02) 100%);
    border: 1px solid var(--accent-border, rgba(96, 165, 250, 0.35));
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  }

  .hero-title-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .title-with-badge {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .app-brand-title {
    font-size: 2.1rem;
    font-weight: 800;
    letter-spacing: -0.02em;
    color: var(--text-primary);
    background: linear-gradient(135deg, var(--text-primary) 30%, var(--accent-light, #60a5fa) 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .version-badge {
    background: var(--accent-badge-bg, rgba(96, 165, 250, 0.15));
    border: 1px solid var(--accent-badge-border, rgba(96, 165, 250, 0.3));
    color: var(--accent-light, #60a5fa);
    font-size: 0.8rem;
    font-weight: 700;
    padding: 3px 10px;
    border-radius: 999px;
  }

  .app-tagline {
    font-size: 0.95rem;
    color: var(--text-secondary);
    font-weight: 500;
  }

  .hero-description {
    font-size: 0.92rem;
    line-height: 1.6;
    color: var(--text-secondary);
    margin-bottom: 20px;
  }

  .hero-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .action-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 0.86rem;
    font-weight: 600;
    padding: 8px 16px;
    border-radius: var(--radius-sm, 6px);
    cursor: pointer;
    transition: all 0.18s ease;
    border: none;
    outline: none;
  }

  .action-btn.primary {
    background: var(--accent-color, #0078d4);
    color: #ffffff;
  }

  .action-btn.primary:hover {
    filter: brightness(1.12);
    transform: translateY(-1px);
  }

  .action-btn.secondary {
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid var(--border-card);
    color: var(--text-primary);
  }

  .action-btn.secondary:hover {
    background: rgba(255, 255, 255, 0.12);
    border-color: var(--border-card-hover);
    transform: translateY(-1px);
  }

  .action-btn.ghost {
    background: transparent;
    border: 1px solid var(--border-subtle);
    color: var(--text-secondary);
  }

  .action-btn.ghost:hover {
    background: rgba(255, 255, 255, 0.05);
    color: var(--text-primary);
  }

  /* Diagnostics Card */
  .diagnostics-card {
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg, 12px);
    padding: 24px 26px;
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .diag-header {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .diag-icon-wrapper {
    width: 38px;
    height: 38px;
    border-radius: 8px;
    background: var(--accent-badge-bg, rgba(255, 255, 255, 0.04));
    border: 1px solid var(--accent-badge-border, rgba(255, 255, 255, 0.12));
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--accent-light, #60a5fa);
  }

  .diag-title {
    font-size: 1.15rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .diag-subtitle {
    font-size: 0.85rem;
    color: var(--text-secondary);
  }

  .specs-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 14px;
  }

  @media (max-width: 680px) {
    .specs-grid {
      grid-template-columns: 1fr;
    }
  }

  .spec-card {
    background: rgba(255, 255, 255, 0.025);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md, 8px);
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    transition: border-color 0.2s;
  }

  .spec-card:hover {
    border-color: var(--border-card-hover);
  }

  .spec-label {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-secondary);
    font-weight: 600;
  }

  .spec-value {
    font-size: 0.94rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .spec-value.highlight {
    color: var(--accent-light, #60a5fa);
  }

  .spec-desc {
    font-size: 0.78rem;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  /* Dev Notice Card */
  .dev-notice-card {
    background: rgba(245, 158, 11, 0.08);
    border: 1px solid rgba(245, 158, 11, 0.25);
    border-radius: var(--radius-md, 10px);
    padding: 18px 22px;
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .notice-icon {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .notice-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .notice-heading {
    font-size: 0.95rem;
    font-weight: 700;
    color: #fbbf24;
  }

  .notice-text {
    font-size: 0.84rem;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .notice-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: rgba(245, 158, 11, 0.18);
    border: 1px solid rgba(245, 158, 11, 0.35);
    color: #fbbf24;
    font-size: 0.82rem;
    font-weight: 600;
    padding: 7px 14px;
    border-radius: 6px;
    cursor: pointer;
    flex-shrink: 0;
    transition: all 0.15s ease;
  }

  .notice-btn:hover {
    background: rgba(245, 158, 11, 0.28);
    transform: translateY(-1px);
  }

  /* Section Tabs Bar */
  .section-tabs-bar {
    display: flex;
    align-items: center;
    gap: 10px;
    border-bottom: 1px solid var(--border-subtle);
    padding-bottom: 12px;
    margin-top: 6px;
  }

  .tab-pill {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: transparent;
    border: 1px solid transparent;
    color: var(--text-secondary);
    font-size: 0.88rem;
    font-weight: 600;
    padding: 8px 16px;
    border-radius: 999px;
    cursor: pointer;
    transition: all 0.18s ease;
  }

  .tab-pill:hover {
    color: var(--text-primary);
    background: rgba(255, 255, 255, 0.05);
  }

  .tab-pill.active {
    background: var(--accent-badge-bg, rgba(96, 165, 250, 0.15));
    border-color: var(--accent-badge-border, rgba(96, 165, 250, 0.35));
    color: var(--accent-light, #60a5fa);
  }

  /* Tab Content */
  .tab-content-panel {
    background: var(--bg-card);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-lg, 12px);
    padding: 28px 30px;
  }

  /* Guide Grid */
  .section-intro {
    margin-bottom: 22px;
  }

  .section-intro h2 {
    font-size: 1.4rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .section-intro p {
    font-size: 0.9rem;
    color: var(--text-secondary);
    margin-top: 4px;
  }

  .guide-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 16px;
  }

  @media (max-width: 720px) {
    .guide-grid {
      grid-template-columns: 1fr;
    }
  }

  .guide-card {
    background: rgba(255, 255, 255, 0.02);
    border: 1px solid var(--border-card);
    border-radius: var(--radius-md, 8px);
    padding: 18px 20px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .guide-card-header {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .guide-card-header h3 {
    font-size: 1rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  :global(.guide-icon) {
    color: var(--accent-light, #60a5fa);
  }

  .guide-card p {
    font-size: 0.85rem;
    color: var(--text-secondary);
    line-height: 1.5;
  }

  .guide-list {
    margin-left: 18px;
    font-size: 0.83rem;
    color: var(--text-secondary);
    line-height: 1.6;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .guide-list strong {
    color: var(--text-primary);
  }

  /* Legal Documents */
  .legal-document {
    display: flex;
    flex-direction: column;
    gap: 22px;
  }

  .legal-header {
    border-bottom: 1px solid var(--border-subtle);
    padding-bottom: 16px;
  }

  .legal-header h2 {
    font-size: 1.45rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .legal-statutory-badge {
    display: inline-block;
    background: rgba(16, 185, 129, 0.12);
    border: 1px solid rgba(16, 185, 129, 0.3);
    color: #10b981;
    font-size: 0.76rem;
    font-weight: 700;
    padding: 2px 10px;
    border-radius: 999px;
    margin-top: 8px;
    margin-bottom: 6px;
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }

  .legal-date {
    font-size: 0.82rem;
    color: var(--text-secondary);
  }

  .legal-body {
    display: flex;
    flex-direction: column;
    gap: 22px;
  }

  .legal-clause {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .legal-clause h4 {
    font-size: 0.94rem;
    font-weight: 700;
    color: var(--text-primary);
    letter-spacing: 0.02em;
  }

  .legal-clause p {
    font-size: 0.86rem;
    color: var(--text-secondary);
    line-height: 1.65;
  }

  .legal-list {
    margin-left: 20px;
    font-size: 0.84rem;
    color: var(--text-secondary);
    line-height: 1.65;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .grievance-card {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid var(--border-card);
    border-radius: 8px;
    padding: 14px 18px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 6px;
  }

  .grievance-title {
    font-size: 0.78rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-secondary);
    font-weight: 600;
  }

  .grievance-name {
    font-size: 1rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .grievance-role, .grievance-jurisdiction {
    font-size: 0.82rem;
    color: var(--text-secondary);
  }

  .grievance-contact {
    font-size: 0.84rem;
    color: var(--accent-light, #60a5fa);
    margin-top: 2px;
  }

  .grievance-contact a {
    color: inherit;
    text-decoration: underline;
  }

  /* Blinking Heart Signature Footer */
  .about-footer-signature {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 30px 20px 10px;
    text-align: center;
  }

  .signature-text {
    font-size: 1.05rem;
    font-weight: 600;
    color: var(--text-primary);
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  .author-name {
    font-weight: 700;
    color: var(--accent-light, #60a5fa);
    letter-spacing: 0.01em;
  }

  .signature-subtext {
    font-size: 0.8rem;
    color: var(--text-secondary);
  }

  /* Blinking Heart CSS Animation */
  .blinking-heart {
    display: inline-block;
    animation: heartBlink 1.25s infinite ease-in-out;
    transform-origin: center center;
    filter: drop-shadow(0 0 6px rgba(255, 51, 102, 0.45));
  }

  @keyframes heartBlink {
    0% {
      transform: scale(1);
      opacity: 1;
    }
    18% {
      transform: scale(1.38);
      opacity: 0.85;
      filter: drop-shadow(0 0 10px rgba(255, 51, 102, 0.8));
    }
    28% {
      transform: scale(1.12);
      opacity: 0.95;
    }
    40% {
      transform: scale(1.32);
      opacity: 0.85;
      filter: drop-shadow(0 0 10px rgba(255, 51, 102, 0.8));
    }
    60% {
      transform: scale(1);
      opacity: 1;
      filter: drop-shadow(0 0 6px rgba(255, 51, 102, 0.45));
    }
    100% {
      transform: scale(1);
      opacity: 1;
    }
  }
</style>
