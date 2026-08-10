import './style.css';
import './app.css';

import {GetSettings as backendGetSettings, SaveSettings as backendSave, Status as backendStatus, Timeline as backendTimeline} from '../wailsjs/go/main/App';
import {EventsOn} from '../wailsjs/runtime/runtime';

document.querySelector('#app').innerHTML = `
    <div class="app-shell">
        <header class="topbar">
            <div class="wordmark"><span class="wordmark-mark">✦</span><span>健康工具箱</span></div>
            <div class="topbar-meta">久坐提醒<span class="topbar-rule"></span>01 <button class="settings-toggle" id="settings-toggle">设置</button></div>
        </header>
        <main>
            <section class="hero" id="hero">
                <div class="hero-copy">
                    <p class="eyebrow">今天也要动一动</p>
                    <h1>工作久一点，<br><span>也要记得起身。</span></h1>
                    <p class="lede">温柔地提醒你离开屏幕，给身体一点空间。</p>
                </div>
                <div class="hero-visual">
                    <div class="track-panel">
                        <div class="track-heading"><span class="track-kicker">当前监控状态</span><strong id="status">待工作</strong></div>
                        <div class="timer" id="elapsed">00:00</div>
                        <div class="track" aria-hidden="true"><span class="track-progress" id="track-progress"></span><span class="track-marker track-marker-start"></span><span class="track-marker track-marker-end"></span></div>
                        <div class="track-meta"><span id="track-start-label">开始</span><span id="track-end-label">起身提醒 / 60 分钟</span></div>
                    </div>
                </div>
            </section>
            <section class="insight-grid">
                <article class="insight-card insight-pink"><span>01</span><strong>有效活动</strong><p>键盘、点击、滚轮和明显的鼠标移动都会被温柔地记住。</p></article>
                <article class="insight-card insight-lavender"><span>02</span><strong>自动恢复</strong><p>启动、闲置暂停和休息结束后，下一次有效活动会自动开始新工作段。</p></article>
                <article class="insight-card insight-teal"><span>03</span><strong>只在本地</strong><p>不记录输入内容、坐标或窗口信息，数据只为这一次提醒服务。</p></article>
            </section>
            <section class="timeline-section" aria-labelledby="timeline-title">
                <div class="section-heading"><div><p class="eyebrow">今天的节奏</p><h2 id="timeline-title">工作与休息记录</h2></div><span class="section-note">仅保留本次运行记录</span></div>
                <div class="timeline-list" id="timeline-list"></div>
            </section>
        </main>
        <footer class="footer"><span>为身体留一点空间</span><span>久坐提醒 · 本地运行</span></footer>
        <div class="settings-backdrop" id="settings-panel" hidden>
            <section class="settings-card" role="dialog" aria-modal="true" aria-labelledby="settings-title">
                <div class="settings-header"><div><p class="eyebrow">偏好设置</p><h2 id="settings-title">提醒与休息</h2></div><button class="close-settings" id="settings-close">×</button></div>
                <p class="settings-description">设置从下一轮工作段或休息期开始生效。</p>
                <label class="setting-label" for="setting-minutes">工作段提醒时长</label>
                <div class="settings-value"><input id="setting-minutes" type="number" min="1" max="180" step="5" value="60"><span>分钟</span></div>
                <input class="settings-range" id="setting-range" type="range" min="1" max="180" step="1" value="60" aria-label="工作段提醒时长">
                <label class="setting-label" for="setting-rest-minutes">提醒后休息时长</label>
                <div class="settings-value"><input id="setting-rest-minutes" type="number" min="1" max="30" step="1" value="3"><span>分钟</span></div>
                <input class="settings-range" id="setting-rest-range" type="range" min="1" max="30" step="1" value="3" aria-label="提醒后休息时长">
                <p class="settings-hint">提醒时长为 1–180 分钟，休息时长为 1–30 分钟。</p>
                <p class="settings-error" id="settings-error" hidden>请输入有效的提醒和休息时长。</p>
                <div class="settings-actions"><button class="button-secondary" id="settings-cancel">取消</button><button class="button-primary" id="settings-save">保存设置</button></div>
            </section>
        </div>
    </div>
`;

const labels = {waiting: '待工作', working: '工作段进行中', 'idle-paused': '闲置暂停', resting: '休息中'};
const timelineLabels = {working: '工作段', resting: '提醒休息', 'idle-paused': '闲置暂停'};
const hasWailsBridge = typeof window.go?.main?.App?.Status === 'function';
const previewSettingsKey = 'health-tool.settings';
let previewState = hasWailsBridge ? 'waiting' : 'working';
let previewElapsed = 0;
let previewReminderMinutes = Number(JSON.parse(localStorage.getItem(previewSettingsKey) || '{}').reminderMinutes) || 60;
let previewRestMinutes = Number(JSON.parse(localStorage.getItem(previewSettingsKey) || '{}').restMinutes) || 3;

const api = {
    status: () => hasWailsBridge ? backendStatus() : Promise.resolve({state: previewState, elapsedSeconds: previewElapsed, reminderMinutes: previewReminderMinutes, restMinutes: previewRestMinutes, restRemainingSeconds: 0}),
    timeline: () => hasWailsBridge ? backendTimeline() : Promise.resolve([]),
    settings: () => hasWailsBridge ? backendGetSettings() : Promise.resolve({reminderMinutes: previewReminderMinutes, restMinutes: previewRestMinutes}),
    saveSettings: (reminderMinutes, restMinutes) => {
        if (hasWailsBridge) return backendSave(reminderMinutes, restMinutes);
        previewReminderMinutes = reminderMinutes;
        previewRestMinutes = restMinutes;
        localStorage.setItem(previewSettingsKey, JSON.stringify({reminderMinutes, restMinutes}));
        return Promise.resolve(true);
    },
};

const heroElement = document.getElementById('hero');
const statusElement = document.getElementById('status');
const elapsedElement = document.getElementById('elapsed');
const progressElement = document.getElementById('track-progress');
const trackStartLabelElement = document.getElementById('track-start-label');
const trackEndLabelElement = document.getElementById('track-end-label');
const settingsPanel = document.getElementById('settings-panel');
const settingMinutes = document.getElementById('setting-minutes');
const settingRange = document.getElementById('setting-range');
const settingRestMinutes = document.getElementById('setting-rest-minutes');
const settingRestRange = document.getElementById('setting-rest-range');
const settingsError = document.getElementById('settings-error');
const timelineList = document.getElementById('timeline-list');

function formatElapsed(seconds) {
    const minutes = Math.floor(seconds / 60).toString().padStart(2, '0');
    const remainder = (seconds % 60).toString().padStart(2, '0');
    return `${minutes}:${remainder}`;
}

function formatTimelineTime(value) {
    return new Intl.DateTimeFormat('zh-CN', {hour: '2-digit', minute: '2-digit'}).format(new Date(value));
}

function formatDuration(seconds) {
    const minutes = Math.floor(seconds / 60);
    const remainder = seconds % 60;
    return minutes ? `${minutes} 分 ${remainder} 秒` : `${remainder} 秒`;
}

function renderTimeline(entries) {
    if (!entries.length) {
        timelineList.innerHTML = '<p class="timeline-empty">还没有工作或休息记录，下一次有效活动会从这里开始。</p>';
        return;
    }
    timelineList.innerHTML = entries.map((entry) => {
        const end = entry.endedAt ? formatTimelineTime(entry.endedAt) : '进行中';
        return `<article class="timeline-entry timeline-${entry.kind}">
            <div class="timeline-dot" aria-hidden="true"></div>
            <div class="timeline-entry-content"><div class="timeline-entry-heading"><strong>${timelineLabels[entry.kind] ?? entry.kind}</strong><span>${formatDuration(entry.durationSeconds)}</span></div>
            <p>${formatTimelineTime(entry.startedAt)} – ${end}</p></div>
        </article>`;
    }).join('');
}

async function refreshStatus() {
    const [current, timeline] = await Promise.all([api.status(), api.timeline()]);
    const resting = current.state === 'resting';
    const totalSeconds = (resting ? current.restMinutes : current.reminderMinutes) * 60;
    const progressSeconds = resting ? totalSeconds - current.restRemainingSeconds : current.elapsedSeconds;
    statusElement.textContent = labels[current.state] ?? current.state;
    elapsedElement.textContent = formatElapsed(resting ? current.restRemainingSeconds : current.elapsedSeconds);
    trackStartLabelElement.textContent = resting ? '休息开始' : '开始';
    trackEndLabelElement.textContent = resting ? `休息结束 / ${current.restMinutes} 分钟` : `起身提醒 / ${current.reminderMinutes} 分钟`;
    heroElement.dataset.state = current.state;
    progressElement.style.width = `${Math.min(Math.max(progressSeconds, 0) / totalSeconds, 1) * 100}%`;
    renderTimeline(timeline);
}

function closeSettings() { settingsPanel.hidden = true; }
async function openSettings() {
    const settings = await api.settings();
    settingMinutes.value = settings.reminderMinutes;
    settingRange.value = settings.reminderMinutes;
    settingRestMinutes.value = settings.restMinutes;
    settingRestRange.value = settings.restMinutes;
    settingsError.hidden = true;
    settingsPanel.hidden = false;
    settingMinutes.focus();
}

document.getElementById('settings-toggle').addEventListener('click', openSettings);
document.getElementById('settings-close').addEventListener('click', closeSettings);
document.getElementById('settings-cancel').addEventListener('click', closeSettings);
settingsPanel.addEventListener('click', (event) => { if (event.target === settingsPanel) closeSettings(); });
settingRange.addEventListener('input', () => { settingMinutes.value = settingRange.value; });
settingMinutes.addEventListener('input', () => { settingRange.value = settingMinutes.value; });
settingRestRange.addEventListener('input', () => { settingRestMinutes.value = settingRestRange.value; });
settingRestMinutes.addEventListener('input', () => { settingRestRange.value = settingRestMinutes.value; });
document.getElementById('settings-save').addEventListener('click', async () => {
    const reminderMinutes = Number(settingMinutes.value);
    const restMinutes = Number(settingRestMinutes.value);
    const validReminder = Number.isInteger(reminderMinutes) && reminderMinutes >= 1 && reminderMinutes <= 180 && (reminderMinutes === 1 || reminderMinutes % 5 === 0);
    const validRest = Number.isInteger(restMinutes) && restMinutes >= 1 && restMinutes <= 30;
    if (!validReminder || !validRest) {
        settingsError.hidden = false;
        return;
    }
    if (await api.saveSettings(reminderMinutes, restMinutes)) {
        closeSettings();
        await refreshStatus();
    }
});

if (window.runtime?.EventsOn) EventsOn('reminder', refreshStatus);
refreshStatus();
window.setInterval(() => {
    if (!hasWailsBridge && previewState === 'working') previewElapsed += 1;
    refreshStatus();
}, 1000);
