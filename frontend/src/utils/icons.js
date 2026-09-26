/**
 * EasyWinGet Intelligent Icon Resolver
 * High-accuracy PRODUCT_MAP & Name heuristics with Google S2 Favicon API & DuckDuckGo fallback.
 */

const PRODUCT_MAP = {
  // Browsers
  'google.chrome': 'google.com',
  'mozilla.firefox': 'mozilla.org',
  'brave.brave': 'brave.com',
  'microsoft.edge': 'microsoft.com',
  'opera.opera': 'opera.com',
  'vivaldi.vivaldi': 'vivaldi.com',
  'torproject.torbrowser': 'torproject.org',

  // Development
  'microsoft.visualstudiocode': 'code.visualstudio.com',
  'microsoft.visualstudio.2022.community': 'visualstudio.microsoft.com',
  'git.git': 'git-scm.com',
  'openjs.nodejs': 'nodejs.org',
  'openjs.nodejs.lts': 'nodejs.org',
  'python.python': 'python.org',
  'python.python.3.11': 'python.org',
  'python.python.3.12': 'python.org',
  'python.python.3.13': 'python.org',
  'docker.dockerdesktop': 'docker.com',
  'kubernetes.kubernetes': 'kubernetes.io',
  'postman.postman': 'postman.com',
  'notepad++.notepad++': 'notepad-plus-plus.org',
  'sublimetext.sublimetext': 'sublimetext.com',
  'jetbrains.intellijideacommunity': 'jetbrains.com',
  'jetbrains.pycharmcommunity': 'jetbrains.com',
  'github.githubdesktop': 'desktop.github.com',
  'insomnia.insomnia': 'insomnia.rest',
  'xpipe.xpipe': 'xpipe.io',
  'jandedobbeleer.ohmyposh': 'ohmyposh.dev',
  'dbeaver.dbeaver': 'dbeaver.io',
  'microsoft.windowsterminal': 'learn.microsoft.com',

  // Media & Creators
  'videolan.vlc': 'videolan.org',
  'spotify.spotify': 'spotify.com',
  'obsproject.obsstudio': 'obsproject.com',
  'audacity.audacity': 'audacityteam.org',
  'blender.blender': 'blender.org',
  'gimp.gimp': 'gimp.org',
  'handbrake.handbrake': 'handbrake.fr',
  'kodi.kodi': 'kodi.tv',
  'foobar2000.foobar2000': 'foobar2000.org',
  'inkscape.inkscape': 'inkscape.org',

  // Communication & Social
  'discord.discord': 'discord.com',
  'telegram.telegramdesktop': 'telegram.org',
  'slack.slack': 'slack.com',
  'whatsapp.whatsapp': 'whatsapp.com',
  'zoom.zoom': 'zoom.us',
  'microsoft.teams': 'teams.microsoft.com',
  'signal.signal': 'signal.org',

  // Utilities & Security
  '7zip.7zip': '7-zip.org',
  'rarlab.winrar': 'win-rar.com',
  'microsoft.powertoys': 'github.com',
  'bleachbit.bleachbit': 'bleachbit.org',
  'ccleaner.ccleaner': 'ccleaner.com',
  'keepassxcteam.keepassxc': 'keepassxc.org',
  'bitwarden.bitwarden': 'bitwarden.com',
  '1password.1password': '1password.com',
  'rufus.rufus': 'rufus.ie',
  'voidtools.everything': 'voidtools.com',
  'crystalmark.crystaldiskinfo': 'crystalmark.info',
  'wireshark.wireshark': 'wireshark.org',
  'oracle.virtualbox': 'virtualbox.org',
  'transmission.transmission': 'transmissionbt.com',
  'qbittorrent.qbittorrent': 'qbittorrent.org',

  // Productivity
  'notion.notion': 'notion.so',
  'obsidian.obsidian': 'obsidian.md',
  'figma.figma': 'figma.com',
  'evernote.evernote': 'evernote.com',
  'anki.anki': 'apps.ankiweb.net',
  'adobe.acrobat.reader.32-bit': 'adobe.com',
  'adobe.acrobat.reader.64-bit': 'adobe.com',
  'adobe.acrobat.pro': 'adobe.com',
  'adobe.creativecloud': 'adobe.com',
  'adobe.photoshop': 'adobe.com',
  'adobe.premiere': 'adobe.com',
  'adobe.illustrator': 'adobe.com',
  'adobe.lightroom': 'adobe.com',
  'adobe.photoshop2025': 'adobe.com',
  'adobe.lightroomclassic': 'adobe.com',
  'adobe.premierepro2025': 'adobe.com',
  'omicronlab.avro': 'omicronlab.com',
  'klocman.bulkcrapuninstaller': 'bcuninstaller.com',
  'bitdefender.bitdefender': 'bitdefender.com',
  'banglaword': 'banglaword.com',
  'google.antigravityide': 'antigravity.google',
  
  // Microsoft Store & Modern Apps
  'xp8jk4hzbvf435': 'autodarkmode.10se1uc.com',
  '9n36ppmp8s23': 'nahimic.com',
  'microsoft.appinstaller': 'microsoft.com',
  'microsoft.desktopappinstaller': 'microsoft.com',
  'microsoft.windowsterminal': 'microsoft.com',
  'spotifyab.spotifymusic': 'spotify.com',
  '5319275a.whatsappdesktop': 'whatsapp.com',
  'chatgpt': 'chatgpt.com',
  'claude': 'anthropic.com',
  'nickjohn.energystarx': 'github.com/Nick-John',
  'flowlauncher.flowlauncher': 'flowlauncher.com',
  'flow-launcher.flow-launcher': 'flowlauncher.com',
  'foxit.foxitreader': 'foxit.com',
  'dbbrowserforsqlite.dbbrowserforsqlite': 'sqlitebrowser.org',
  'eclipseadoptium.temurin.21.jdk': 'adoptium.net',
  'cloudflare.warp': 'cloudflare.com',
  'cloudflare.cloudflared': 'cloudflare.com',

  // Gaming
  'valve.steam': 'store.steampowered.com',
  'epicgames.epicgameslauncher': 'store.epicgames.com',
  'gog.galaxy': 'gog.com',
  'electronicarts.eaapp': 'ea.com',
  'ubisoft.connect': 'ubisoft.com'
};

const COMMON_APP_NAMES = [
  { match: ['chrome'], domain: 'google.com' },
  { match: ['firefox'], domain: 'mozilla.org' },
  { match: ['edge'], domain: 'microsoft.com' },
  { match: ['brave'], domain: 'brave.com' },
  { match: ['opera'], domain: 'opera.com' },
  { match: ['visual studio code', 'vscode', 'code - oss'], domain: 'code.visualstudio.com' },
  { match: ['visual studio'], domain: 'visualstudio.microsoft.com' },
  { match: ['discord'], domain: 'discord.com' },
  { match: ['spotify'], domain: 'spotify.com' },
  { match: ['vlc'], domain: 'videolan.org' },
  { match: ['git '], domain: 'git-scm.com' },
  { match: ['node.js', 'nodejs'], domain: 'nodejs.org' },
  { match: ['python'], domain: 'python.org' },
  { match: ['docker'], domain: 'docker.com' },
  { match: ['postman'], domain: 'postman.com' },
  { match: ['7-zip', '7zip'], domain: '7-zip.org' },
  { match: ['winrar'], domain: 'win-rar.com' },
  { match: ['powertoys'], domain: 'github.com' },
  { match: ['notepad++'], domain: 'notepad-plus-plus.org' },
  { match: ['sublime'], domain: 'sublimetext.com' },
  { match: ['obs studio', 'obs '], domain: 'obsproject.com' },
  { match: ['steam'], domain: 'store.steampowered.com' },
  { match: ['whatsapp'], domain: 'whatsapp.com' },
  { match: ['telegram'], domain: 'telegram.org' },
  { match: ['bitwarden'], domain: 'bitwarden.com' },
  { match: ['notion'], domain: 'notion.so' },
  { match: ['obsidian'], domain: 'obsidian.md' },
  { match: ['figma'], domain: 'figma.com' },
  { match: ['zoom'], domain: 'zoom.us' },
  { match: ['teams'], domain: 'teams.microsoft.com' },
  { match: ['blender'], domain: 'blender.org' },
  { match: ['gimp'], domain: 'gimp.org' },
  { match: ['audacity'], domain: 'audacityteam.org' },
  { match: ['handbrake'], domain: 'handbrake.fr' },
  { match: ['dbeaver'], domain: 'dbeaver.io' },
  { match: ['wireshark'], domain: 'wireshark.org' },
  { match: ['virtualbox'], domain: 'virtualbox.org' },
  { match: ['qbittorrent'], domain: 'qbittorrent.org' },
  { match: ['everything'], domain: 'voidtools.com' },
  { match: ['rufus'], domain: 'rufus.ie' },
  { match: ['avro'], domain: 'omicronlab.com' },
  { match: ['bulk crap uninstaller', 'bcuninstaller'], domain: 'bcuninstaller.com' },
  { match: ['bitdefender'], domain: 'bitdefender.com' },
  { match: ['banglaword'], domain: 'banglaword.com' },
  { match: ['photoshop', 'lightroom', 'premiere'], domain: 'adobe.com' },
  { match: ['auto dark mode', 'autodarkmode'], domain: 'autodarkmode.10se1uc.com' },
  { match: ['chatgpt'], domain: 'chatgpt.com' },
  { match: ['claude'], domain: 'anthropic.com' },
  { match: ['nahimic'], domain: 'nahimic.com' },
  { match: ['app installer'], domain: 'microsoft.com' },
  { match: ['energy star'], domain: 'github.com' },
  { match: ['flow launcher'], domain: 'flowlauncher.com' },
  { match: ['foxit'], domain: 'foxit.com' },
  { match: ['sqlite'], domain: 'sqlitebrowser.org' }
];

function sanitizeDomain(raw) {
  if (!raw) return null;
  let d = raw.trim().toLowerCase();
  d = d.replace(/^https?:\/\//i, '');
  d = d.split('/')[0].split('?')[0].split(':')[0].trim();
  if (d.startsWith('www.')) d = d.substring(4);
  return d || null;
}

/**
 * Extracts a clean domain from package ID, homepage, or app name
 */
export function getDomainForPackage(packageId, homepage = '', name = '') {
  // 1. Homepage has highest fidelity
  if (homepage) {
    try {
      const u = new URL(homepage.startsWith('http') ? homepage : `https://${homepage}`);
      const clean = sanitizeDomain(u.hostname);
      if (clean) return clean;
    } catch (e) {}
  }

  const lowerId = (packageId || '').toLowerCase().replace(/[^a-z0-9.]/g, '').trim();

  // 2. Direct match in Product Map
  if (lowerId && PRODUCT_MAP[lowerId]) {
    return sanitizeDomain(PRODUCT_MAP[lowerId]);
  }

  // 3. Name-based heuristics
  const cleanName = (name || '').toLowerCase().trim();
  if (cleanName) {
    for (const item of COMMON_APP_NAMES) {
      if (item.match.some(m => cleanName.includes(m))) {
        return item.domain;
      }
    }
  }

  // 4. Microsoft Store 12-character ID
  if (/^[a-zA-Z0-9]{12}$/.test(packageId)) {
    return 'apps.microsoft.com';
  }

  // 5. Filter out installer registry GUIDs
  if (
    lowerId.startsWith('arp') ||
    lowerId.startsWith('msix') ||
    lowerId.startsWith('user') ||
    lowerId.startsWith('machine') ||
    lowerId.includes('{')
  ) {
    return null;
  }

  // 6. Partial match in Product Map
  if (lowerId) {
    for (const [key, domain] of Object.entries(PRODUCT_MAP)) {
      if (lowerId.includes(key) || key.includes(lowerId)) {
        return sanitizeDomain(domain);
      }
    }

    // 7. Publisher.Product heuristic
    const parts = lowerId.split('.');
    if (parts.length >= 2) {
      const prod = parts[1];
      if (prod.length > 2) {
        return `${prod}.com`;
      }
      const pub = parts[0];
      if (pub.length > 2) {
        return `${pub}.com`;
      }
    }
  }

  return null;
}

/**
 * Returns Google S2 high-res favicon URL for a domain
 */
export function getGoogleIconUrl(domain) {
  const clean = sanitizeDomain(domain);
  if (!clean) return null;
  return `https://www.google.com/s2/favicons?domain=${encodeURIComponent(clean)}&sz=128`;
}

/**
 * Returns DuckDuckGo icon URL as secondary fallback
 */
export function getDuckDuckGoIconUrl(domain) {
  const clean = sanitizeDomain(domain);
  if (!clean) return null;
  return `https://icons.duckduckgo.com/ip3/${encodeURIComponent(clean)}.ico`;
}

/**
 * Returns the primary Google S2 icon URL for a package
 */
export function getIconUrl(packageId, homepage = '', name = '') {
  const domain = getDomainForPackage(packageId, homepage, name);
  return getGoogleIconUrl(domain);
}
