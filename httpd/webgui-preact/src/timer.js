/**
 * @param {Timer} timer - timer to post to
 * @param {string} endpoint - anything to be added to /api/timer. must start with "/"
*/
export function postTimer(timer, endpoint="") {
  let xhr = new XMLHttpRequest();
  xhr.open("POST", '/api/timer' + endpoint, true);
  xhr.setRequestHeader("Content-Type", "application/json; charset=UTF-8")
  xhr.responseType = 'json'
  xhr.send(JSON.stringify(timer));
}


/**
 * @param {TimerMode} mode
 */
export function timerModeString(mode) {
  switch (mode) {
    case 0:
      return "Pomodoro"
    case 1:
      return "Short Break"
    case 2:
      return "Long Break"
  }
}

/**
 * @typedef {number} Duration
 * */

/**
 * @typedef {number} TimerMode
 * */

/**
 * @typedef {Object} TimerConfig
 * @property {Duration[]} Duration - default duration of each mode (an array of 3)
 * @property {number} Sessions - count of sessions per timer
 * @property {Boolean} Paused - would timer be paused at the start of a timer?
 */

/**
 * @typedef {Object} Timer
 * @property {Duration} Duration - current duration left of timer
 * @property {TimerConfig} Config - config of the timer
 * @property {TimerMode} Mode - current timer mode
 * @property {Boolean} Paused - is timer paused right now?
 * @property {number} FinishedSessions - number of finished sessions
 */

