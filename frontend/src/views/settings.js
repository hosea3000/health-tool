import {CheckForUpdates as backendCheck, CurrentVersion as backendVersion} from '../../wailsjs/go/main/App';
import {BrowserOpenURL} from '../../wailsjs/runtime/runtime';

const hasWailsBridge = typeof window.go?.main?.App?.CheckForUpdates === 'function';

const api = {
    version: () => (hasWailsBridge ? backendVersion() : Promise.resolve('dev')),
    check: () => (hasWailsBridge ? backendCheck() : Promise.resolve({status: 'up-to-date', currentVersion: 'dev', latestVersion: '', releaseUrl: '', message: '开发预览，不检查更新'})),
};

function escapeHtml(value) {
    return String(value).replace(/[&<>"']/g, (c) => ({'&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'}[c]));
}

export function renderSettings({onBack}) {
    const app = document.getElementById('app');
    app.innerHTML = `
        <div class="app-shell">
            <header class="topbar">
                <div class="topbar-left">
                    <button class="wordmark wordmark-link" id="back-button" aria-label="返回首页"><span class="wordmark-mark">✦</span><span>健康工具箱</span></button>
                </div>
                <div class="topbar-meta"><span>设置</span></div>
            </header>
            <main class="settings-page">
                <section class="settings-section">
                    <p class="eyebrow">关于</p>
                    <div class="settings-block">
                        <strong class="settings-app-name">健康工具箱</strong>
                        <p class="settings-app-version" id="app-version">版本加载中…</p>
                        <p class="settings-app-desc">为身体留一点空间 · 本地运行</p>
                    </div>
                </section>
                <section class="settings-section">
                    <p class="eyebrow">更新</p>
                    <div class="settings-block">
                        <p class="settings-app-desc">检查 GitHub Release 上是否有新版本，发现新版本后可前往查看</p>
                        <button class="button-primary settings-check-btn" id="check-updates">检查更新</button>
                        <div class="update-feedback" id="update-feedback" hidden></div>
                    </div>
                </section>
            </main>
            <footer class="footer"><span>为身体留一点空间</span><span>设置 · 本地运行</span></footer>
        </div>
    `;

    document.getElementById('back-button').addEventListener('click', onBack);

    const versionNode = document.getElementById('app-version');
    api.version().then((v) => {
        versionNode.textContent = v === 'dev' ? '开发版本' : `版本 v${v}`;
    }).catch(() => {
        versionNode.textContent = '版本未知';
    });

    const checkBtn = document.getElementById('check-updates');
    const feedback = document.getElementById('update-feedback');

    function renderFeedback(result) {
        feedback.hidden = false;
        if (result.status === 'update-available') {
            feedback.innerHTML = `
                <p class="update-feedback-text">发现新版本 v${escapeHtml(result.latestVersion)}</p>
                <button class="button-primary settings-check-btn" id="go-release">前往 GitHub 查看</button>
            `;
            document.getElementById('go-release').addEventListener('click', () => {
                const url = result.releaseUrl;
                if (!url) return;
                if (hasWailsBridge) BrowserOpenURL(url);
                else window.open(url, '_blank');
            });
            return;
        }
        const fallback = result.message || (result.status === 'up-to-date' ? '已是最新版本' : '检查更新失败');
        feedback.innerHTML = `<p class="update-feedback-text">${escapeHtml(fallback)}</p>`;
    }

    checkBtn.addEventListener('click', async () => {
        checkBtn.disabled = true;
        checkBtn.textContent = '检查中…';
        feedback.hidden = true;
        try {
            renderFeedback(await api.check());
        } catch (err) {
            feedback.hidden = false;
            feedback.innerHTML = `<p class="update-feedback-text update-feedback-error">检查更新失败，请稍后重试</p>`;
        } finally {
            checkBtn.disabled = false;
            checkBtn.textContent = '检查更新';
        }
    });

    return {
        destroy() {},
        onReminder() {},
    };
}