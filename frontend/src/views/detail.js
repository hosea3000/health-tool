import {registry} from '../tools/index.js';

export function renderDetail(toolId, {onBack}) {
    const tool = registry[toolId];
    if (!tool) return {destroy() {}, onReminder() {}};
    const app = document.getElementById('app');
    app.innerHTML = `
        <div class="app-shell">
            <header class="topbar">
                <div class="topbar-left">
                    <button class="back-button" id="back-button">‹ 返回</button>
                    <div class="wordmark"><span class="wordmark-mark">✦</span><span>健康工具箱</span></div>
                </div>
                <div class="topbar-meta"><span>${tool.name}</span></div>
            </header>
            <main id="tool-content"></main>
            <footer class="footer"><span>为身体留一点空间</span><span>${tool.name} · 本地运行</span></footer>
        </div>
    `;

    const host = document.getElementById('tool-content');
    const instance = tool.renderDetail(host);
    document.getElementById('back-button').addEventListener('click', onBack);

    return {
        destroy() { instance?.destroy?.(); },
        onReminder() { instance?.tick?.(); },
    };
}
