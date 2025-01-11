import { exec } from 'child_process';

exec('amixer -D pulse set Capture toggle',
  (err, stdout, stderr) => {
    if (err) console.log(err);
    if (stdout.includes('[on]')) console.log('Unmuted');
    if (stdout.includes('[off]')) console.log('Muted');
    if (stderr) console.log(stderr);
  }
);