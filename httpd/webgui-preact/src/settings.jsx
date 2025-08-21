import { useEffect, useMemo, useState } from 'preact/hooks';
import { Radio, Button, parseDuration, formatDuration } from "./utils"
import { postTimer, timerModeString } from "./timer"
import { sendNotification } from "./utils"

  function updateNotifications() {
    Notification.requestPermission((result) => {
      if (result === "granted") {
        sendNotification("Notification enabled!")
      } else {
        setNotification(false);
      }
    });
  }
export function Settings(p) {
  const duration = useMemo(() => [0, 1, 2].map(mode => formatDuration(p.timer.Config.Duration[mode])), [p.timer.Config.Duration])
  const [submitValue, setSubmitValue] = useState("save")
  

  return (
    <div id="settings-wrapper" class={"z-100 top-0 right-0 absolute flex min-w-full min-h-screen overflow-hidden" + (p.hidden ? " hidden" : "")}>
      <div id="settings-overlay" class="transition ease-in-out duration-300 grow bg-black/30" onClick={p.onClose} />
      <div id="settings" class="p-4 overflow-y-scroll transition-all bg-white dark:bg-zinc-800 rounded-l-lg shadow-md hover:shadow-lg ease-in-out duration-300 float-right h-screen text-wrap w-full md:w-70">
        <div class="flex justify-end">
          <Button onClick={p.onClose} title="close settings">{close_icon}</Button>
        </div>
        <form onSubmit={(e) => {
          e.preventDefault();
          postTimer(p.timer) ;
          setSubmitValue(`saved! reset timer`)
          setTimeout(() => {
            setSubmitValue("save")
          }, 3000);
        }} class="flex flex-col gap-4">
          {[0, 1, 2].map((mode) =>
            <div>
              <label htmlFor={`timer-config-duration-${mode}`}>{timerModeString(mode)} duration</label>
              <input id={`timer-config-duration-${mode}`}
                class="rounded p-2 text-md bg-zinc-200 dark:bg-zinc-700 w-full"
                type="text" value={duration[mode]}
                onChange={(e) => p.timer.Config.Duration[mode] = parseDuration(e.target.value)}
              />
            </div>
          )}
          <div>
            <label htmlFor="timer-config-sessions">Sessions</label>
            <input id="timer-config-sessions"
              class="rounded p-2 text-md bg-zinc-200 dark:bg-zinc-700 w-full"
              type="text" value={p.timer.Config.Sessions}
              onChange={(e) => p.timer.Config.Sessions = parseInt(e.target.value)}
            />
          </div>
          <Radio id="timer-config-paused" checked={p.timer.Config.Paused} onChange={() => p.timer.Config.Paused = !p.timer.Config.Paused}>
            is timer initially paused
          </Radio>
          <Radio id="webgui-notification" checked={p.notification} onChange={(e) => {
            p.setNotification(!p.notification);
            if (!p.notification) {
              updateNotifications();
            }
          }}>
            send notifications
          </Radio>
          <input type="submit" value={submitValue} class="cursor-pointer p-2 rounded transition ease-in-out duration-300 dark:bg-zinc-900 dark:hover:text-zinc-300 hover:text-zinc-700 bg-zinc-200" />
        </form>
      </div>
    </div>
  );
}

const close_icon = <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" class="size-6">
  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18 18 6M6 6l12 12" />
</svg>

