import http from 'k6/http';
import { sleep } from 'k6';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';

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
  };
}
