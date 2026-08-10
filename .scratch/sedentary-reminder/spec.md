---
title: "实现久坐提醒工具"
labels:
  - ready-for-agent
---

## Problem Statement

用户在电脑前持续工作时容易忘记起身活动。用户需要一个 Windows 桌面工具，根据真实的键盘和鼠标操作判断是否正在工作，并在连续工作达到 60 分钟时提醒起身；离开电脑、锁屏或睡眠期间不应继续计入工作段。

## Solution

提供一个由用户手动启动的 Windows 桌面工具。工具只在本地内存中观察有效活动的发生时间和类型，不记录按键内容、鼠标坐标、窗口标题，也不上传数据。

工具以明确的状态管理工作：启动或继续后进入待工作，首次有效活动创建工作段；工作段达到 60 分钟后发送一次 Windows 通知并进入提醒暂停；连续 5 分钟没有有效活动后进入闲置暂停。暂停状态不会因新的输入自动恢复，用户必须通过系统托盘或通知中的继续操作恢复，恢复后创建一个从 0 分钟开始的新工作段。

## User Stories

1. As a Windows desktop user, I want to start the tool manually, so that monitoring only runs when I choose to use it.
2. As a user, I want the tool to enter a待工作 state after startup, so that merely opening the tool does not count as working time.
3. As a user, I want the first有效活动 to start a工作段, so that the timer reflects actual computer use.
4. As a user, I want keyboard key presses to count as有效活动, so that typing keeps my work state current.
5. As a user, I want mouse clicks to count as有效活动, so that ordinary pointer interaction keeps my work state current.
6. As a user, I want mouse-wheel actions to count as有效活动, so that scrolling is treated as active computer use.
7. As a user, I want明显的鼠标移动 to count as有效活动, so that pointer-based work is recognized.
8. As a user, I want tiny mouse jitter to be ignored, so that hardware noise cannot keep the tool active while I am away.
9. As a user, I want the work段 to remain active while valid activity continues, so that short gaps during reading or thinking do not immediately stop the timer.
10. As a user, I want five minutes without有效活动 to enter闲置暂停, so that time away from the computer is excluded from the work段.
11. As a user, I want the tool to stop counting during lock, sleep, or other periods without input, so that system idle time cannot trigger a false reminder.
12. As a user, I want a notification when a工作段 reaches 60 minutes, so that I receive a clear prompt to stand up.
13. As a user, I want each工作段 to trigger at most one sedentary reminder, so that the tool does not repeatedly interrupt the same work段.
14. As a user, I want the reminder to enter提醒暂停, so that the timer cannot silently continue after I have been told to take a break.
15. As a user, I want new keyboard or mouse input to leave a pause unchanged, so that resuming monitoring is always an explicit choice.
16. As a user, I want to continue from the system tray, so that I can resume even after the notification is dismissed.
17. As a user, I want to continue from the notification, so that I can resume quickly when I return from a break.
18. As a user, I want continuing to create a new工作段 starting at zero, so that the previous work段 cannot carry time across a break.
19. As a user, I want to pause or resume monitoring from the system tray, so that I can control the tool without opening a main window.
20. As a user, I want the tool to remain quiet when no work段 has started, so that an unattended launch cannot produce a reminder.
21. As a user, I want the tool to avoid collecting input content, coordinates, or window information, so that monitoring does not expose private activity.
22. As a user, I want all monitoring data to stay local and transient, so that the tool has no remote-data or history-retention risk.
23. As a user, I want the tool to provide understandable tray state, so that I can tell whether it is待工作, working,闲置暂停, or提醒暂停.
24. As a user, I want the notification and tray actions to be safe when invoked more than once, so that duplicate clicks cannot create duplicate工作段 or reminders.

## Implementation Decisions

- The first release targets Windows desktop only.
- Monitoring starts when the user launches the tool manually. Automatic Windows startup is out of scope for the first release.
- The core behavior is a small state machine with待工作, active工作段,闲置暂停, and提醒暂停 states. The state machine owns transitions and elapsed-time rules; input capture, clock access, tray UI, and Windows notification delivery remain adapters around it.
- Startup and continue enter待工作. The first有效活动 transitions待工作 to a new工作段 and establishes its start time.
-有效活动 consists of keyboard key presses, mouse clicks, mouse-wheel actions, and明显的鼠标移动. Tiny movement below the chosen movement threshold is ignored.
- A工作段 is eligible for a reminder after 60 minutes of continuous active time. A continuous period is broken by five minutes without有效活动; the inactive interval is never counted as work time.
- Five minutes without有效活动 transitions an active工作段 to闲置暂停. Lock and sleep are treated as non-active conditions and must not advance work time; on return the user still explicitly continues monitoring.
- Reaching 60 minutes sends one Windows notification and transitions to提醒暂停. Further input cannot advance time or implicitly resume monitoring.
- Continue is available from the system tray and the notification action. Either action creates a new工作段 with elapsed time reset to zero.
- The tray is the persistent control surface for current state and manual pause/continue actions. The tool does not require a main window for normal operation.
- The monitoring boundary exposes only event kind and event time to the domain logic. No key value, text, mouse coordinate, active-window identity, or remote telemetry is collected.
- No historical activity or reminder data is persisted in the first release.
- Notification delivery and tray actions are best-effort integrations; the domain state must remain correct and idempotent if an integration callback is duplicated.

## Testing Decisions

- Tests will assert externally observable state transitions, elapsed time, reminder decisions, and available actions. They will not assert hook-library calls, timer implementation details, or Windows API internals.
- The highest useful seam is the domain state boundary with an injected clock and synthetic有效活动 events. This permits deterministic coverage of startup, first input, five-minute inactivity, 60-minute reminder, continue, duplicate continue, and late events.
- The domain boundary will be tested for keyboard, click, wheel, significant movement, and ignored tiny movement.
- Boundary tests will cover activity immediately before the five-minute cutoff, exactly at the cutoff, immediately before 60 minutes, exactly at 60 minutes, and input received after a pause.
- Integration tests will cover the adapter contract that a reminder produces one notification request and that tray/notification continue produces one domain continue action. They will avoid requiring a live desktop notification in the default test suite.
- A small end-to-end smoke check may verify that the Windows process can start, expose a tray control, and shut down cleanly, but it is not the primary correctness test.
- There is no existing codebase testing prior art: the repository currently contains only the domain glossary, so the first tests should establish the single deterministic domain seam rather than introduce a broad test framework or UI test suite.

## Out of Scope

- macOS, Linux, or cross-platform support.
- Automatic launch with Windows.
- Work-app or active-window detection.
- Recording keystrokes, mouse coordinates, window titles, screenshots, or detailed productivity analytics.
- Cloud synchronization, accounts, remote telemetry, or historical reports.
- Custom schedules, working hours, multiple reminder intervals, snooze behavior, or reminder sounds.
- Automatic resume after input, unlock, or wake.
- Multiple concurrent users or per-user profiles.
- A full settings UI; the 60-minute and five-minute thresholds are fixed in the first release.

## Further Notes

- The domain glossary in `CONTEXT.md` is authoritative for待工作,工作段,有效活动,闲置暂停,提醒暂停, and继续.
- The local issue tracker is used because this directory has no git remote or configured external issue tracker. The issue is marked `ready-for-agent` as requested.
- The movement threshold is a product behavior that must be selected and tested consistently; it should not be inferred from raw hardware jitter.
