import {CheckForUpdates as backendCheck, CurrentVersion as backendVersion, DownloadAndApplyUpdate as backendDownload, ApplyUpdateAndRestart as backendRestart, PendingUpdateInfo as backendPendingUpdate} from '../../wailsjs/go/main/App';
import {BrowserOpenURL, EventsOn, EventsOff} from '../../wailsjs/runtime/runtime';

const hasWailsBridge = typeof window.go?.main?.App?.CheckForUpdates === 'function';

const UPDATE_PROGRESS_EVENT = 'update:progress';

const api = {
    version: () => (hasWailsBridge ? backendVersion() : Promise.resolve('dev')),
    check: () => (hasWailsBridge ? backendCheck() : Promise.resolve({status: 'up-to-date', currentVersion: 'dev', latestVersion: '', downloadUrl: '', releaseUrl: '', message: '开发预览，不检查更新'})),
    download: () => (hasWailsBridge ? backendDownload() : Promise.resolve('开发预览，不支持自动更新')),
    restart: () => (hasWailsBridge ? backendRestart() : Promise.resolve('开发预览，不支持自动更新')),
    pendingUpdate: () => (hasWailsBridge ? backendPendingUpdate() : Promise.resolve({exists: false, version: ''})),
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
                        <p class="settings-app-desc">检查 GitHub Release 上是否有新版本，发现新版本后可一键下载并重启生效</p>
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
    // 更新流程状态：idle / downloading / downloaded（待重启）/ restarting
    let updateState = 'idle';
    let applyBtn = null;
    let progressBox = null;
    let progressFill = null;
    let progressText = null;

    // 展示主反馈区（检查结果或错误文案）
    function showFeedback(html) {
        feedback.hidden = false;
        feedback.innerHTML = html;
    }

    function hideFeedback() {
        feedback.hidden = true;
        feedback.innerHTML = '';
    }

    // 构建更新按钮区：label 为按钮文案，onClick 为点击行为
    function renderApplyAction(label, onClick) {
        feedback.innerHTML = `
            <div class="update-actions">
                <button class="button-primary settings-check-btn" id="apply-update">${escapeHtml(label)}</button>
                <button class="button-secondary settings-check-btn" id="go-release">前往 GitHub 查看</button>
            </div>
            <div class="update-progress" id="update-progress" hidden>
                <div class="update-progress-track"><div class="update-progress-fill" id="update-progress-fill"></div></div>
                <p class="update-feedback-text" id="update-progress-text">下载中 0%</p>
            </div>
        `;
        applyBtn = document.getElementById('apply-update');
        applyBtn.addEventListener('click', onClick);
        progressBox = document.getElementById('update-progress');
        progressFill = document.getElementById('update-progress-fill');
        progressText = document.getElementById('update-progress-text');
        document.getElementById('go-release').addEventListener('click', () => {
            const url = document.getElementById('update-feedback').dataset.releaseUrl;
            if (!url) return;
            if (hasWailsBridge) BrowserOpenURL(url);
            else window.open(url, '_blank');
        });
    }

    // 进入下载态：显示进度条、禁用按钮
    function enterDownloading() {
        updateState = 'downloading';
        applyBtn.disabled = true;
        applyBtn.textContent = '下载中…';
        progressBox.hidden = false;
        progressFill.style.width = '0%';
        progressText.textContent = '下载中 0%';
    }

    function showProgress(percent, text) {
        progressFill.style.width = `${Math.max(0, Math.min(100, percent))}%`;
        progressText.textContent = text;
    }

    function handleProgressEvent(event) {
        if (updateState !== 'downloading' && updateState !== 'downloaded') return;
        switch (event.phase) {
            case 'downloading':
                showProgress(event.percent || 0, `下载中 ${event.percent || 0}%`);
                break;
            case 'completed':
                showProgress(100, '下载完成，等待确认…');
                break;
            case 'cancelled':
                // 用户选择稍后重启：按钮切换为「重启更新」
                updateState = 'downloaded';
                progressBox.hidden = true;
                showFeedback(`
                    <p class="update-feedback-text">新版本已下载完成，可随时重启生效</p>
                    <div class="update-actions">
                        <button class="button-primary settings-check-btn" id="apply-update">重启更新</button>
                        <button class="button-secondary settings-check-btn" id="go-release">前往 GitHub 查看</button>
                    </div>
                `);
                applyBtn = document.getElementById('apply-update');
                applyBtn.addEventListener('click', onRestartClick);
                document.getElementById('go-release').addEventListener('click', onGoReleaseClick);
                break;
            case 'error':
                updateState = 'idle';
                applyBtn.disabled = false;
                applyBtn.textContent = '立即更新';
                progressBox.hidden = false;
                progressFill.style.width = '0%';
                progressText.textContent = event.message || '更新失败，请稍后重试';
                break;
        }
    }

    function onGoReleaseClick() {
        const url = document.getElementById('update-feedback').dataset.releaseUrl;
        if (!url) return;
        if (hasWailsBridge) BrowserOpenURL(url);
        else window.open(url, '_blank');
    }

    async function onDownloadClick() {
        applyBtn.disabled = true;
        try {
            const msg = await api.download();
            if (msg) {
                showFeedback(`<p class="update-feedback-text update-feedback-error">${escapeHtml(msg)}</p>`);
                applyBtn.disabled = false;
                return;
            }
            enterDownloading();
        } catch {
            showFeedback('<p class="update-feedback-text update-feedback-error">启动更新失败，请稍后重试</p>');
            applyBtn.disabled = false;
        }
    }

    async function onRestartClick() {
        applyBtn.disabled = true;
        applyBtn.textContent = '正在重启…';
        try {
            const msg = await api.restart();
            if (msg === '已取消') {
                // 用户再次选择稍后，保持「重启更新」可用
                applyBtn.disabled = false;
                applyBtn.textContent = '重启更新';
                return;
            }
            if (msg) {
                updateState = 'idle';
                showFeedback(`<p class="update-feedback-text update-feedback-error">${escapeHtml(msg)}</p>`);
                applyBtn.textContent = '立即更新';
                applyBtn.disabled = false;
                return;
            }
            // 空串：已进入退出重启流程，等待应用退出
            updateState = 'restarting';
            applyBtn.textContent = '即将重启…';
        } catch {
            applyBtn.disabled = false;
            applyBtn.textContent = '重启更新';
        }
    }

    // 渲染「重启更新」入口（.new 已就绪，版本号来自后端元数据）
    function renderPendingRestart(version) {
        updateState = 'downloaded';
        feedback.hidden = false;
        feedback.dataset.releaseUrl = '';
        feedback.innerHTML = `
            <p class="update-feedback-text">新版本 ${version ? `v${escapeHtml(version)}` : ''} 已下载完成，可随时重启生效</p>
            <div class="update-actions">
                <button class="button-primary settings-check-btn" id="apply-update">重启更新</button>
                <button class="button-secondary settings-check-btn" id="go-release">前往 GitHub 查看</button>
            </div>
        `;
        applyBtn = document.getElementById('apply-update');
        applyBtn.addEventListener('click', onRestartClick);
        document.getElementById('go-release').addEventListener('click', onGoReleaseClick);
    }

    async function renderFeedback(result) {
        feedback.hidden = false;
        feedback.dataset.releaseUrl = result.releaseUrl || '';
        if (result.status === 'update-available') {
            if (result.downloadUrl) {
                // 有资产下载地址：先查待更新状态，.new 已就绪则给「重启更新」，否则给「立即更新」
                let pending = null;
                try {
                    pending = await api.pendingUpdate();
                } catch {
                    pending = null;
                }
                if (pending && pending.exists) {
                    renderPendingRestart(pending.version);
                } else {
                    renderApplyAction('立即更新', onDownloadClick);
                    feedback.insertAdjacentHTML('afterbegin', `<p class="update-feedback-text">发现新版本 v${escapeHtml(result.latestVersion)}</p>`);
                }
            } else {
                // 资产缺失：仅提供跳转
                showFeedback(`
                    <p class="update-feedback-text">发现新版本 v${escapeHtml(result.latestVersion)}</p>
                    <button class="button-primary settings-check-btn" id="go-release">前往 GitHub 查看</button>
                `);
                document.getElementById('go-release').addEventListener('click', onGoReleaseClick);
            }
            return;
        }
        const fallback = result.message || (result.status === 'up-to-date' ? '已是最新版本' : '检查更新失败');
        showFeedback(`<p class="update-feedback-text">${escapeHtml(fallback)}</p>`);
    }

    checkBtn.addEventListener('click', async () => {
        checkBtn.disabled = true;
        checkBtn.textContent = '检查中…';
        feedback.hidden = true;
        updateState = 'idle';
        try {
            await renderFeedback(await api.check());
        } catch (err) {
            showFeedback('<p class="update-feedback-text update-feedback-error">检查更新失败，请稍后重试</p>');
        } finally {
            checkBtn.disabled = false;
            checkBtn.textContent = '检查更新';
        }
    });

    // 进入设置页即查询待更新状态：.new 已就绪（上次取消后遗留）时直接显示「重启更新」
    api.pendingUpdate().then((info) => {
        if (info && info.exists) {
            renderPendingRestart(info.version);
        }
    }).catch(() => {});

    // 订阅下载进度事件（Go 端通过 update:progress 推送）
    const unsubscribe = EventsOn(UPDATE_PROGRESS_EVENT, (event) => {
        handleProgressEvent(event);
    });

    return {
        destroy() {
            unsubscribe();
        },
        onReminder() {},
    };
}
