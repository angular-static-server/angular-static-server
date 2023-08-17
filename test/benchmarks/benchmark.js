import http from 'k6/http';
import { sleep } from 'k6';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';
import { textSummary } from 'https://jslib.k6.io/k6-summary/0.0.3/index.js';

export const options = {
  vus: 100,
  duration: '30s',
};

export default function () {
  http.get('http://localhost:8080/de-CH/');
  sleep(1);
}

export function handleSummary(data) {
  return {
    [`k6-${__ENV.TYPE || ''}-report.html`]: htmlReport(data),
    [`k6-${__ENV.TYPE || ''}-report.txt`]: textSummary(data, { indent: '→', enableColors: false }),
  };
}
