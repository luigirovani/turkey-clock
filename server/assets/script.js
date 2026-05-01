
const data = window.__NTP_DATA__;
let currentTimestamp = data ? Date.parse(data.current_time) : Date.now();
let serverError = true;
const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));

function formatLocal() {
    const d = new Date(currentTimestamp);
    const formatted = new Intl.DateTimeFormat(undefined, {
        dateStyle: 'short',
        timeStyle: 'medium'
    }).format(d);
    document.getElementById('clock').innerText = formatted;
}

async function syncWithNTP() {
    const urlParams = new URLSearchParams(window.location.search);
    urlParams.set('precision_unit', 'auto');
    const params = `?${urlParams.toString()}`;

    try {
        const start = performance.now();
        const res = await fetch('/time' + params);
        const data = await res.json();
        if (data.error) throw new Error();

        const end = performance.now();
        const requestTime = end - start;

        currentTimestamp = Date.parse(data.current_time) + requestTime;
        
        if (serverError) {
            document.getElementById('main-card').classList.remove('error-state');
            document.getElementById('turkey-icon').src = "https://aecrypto.io/static/cartoons/turkey_smiling.png";
            document.getElementById('status-label').innerText = "NTP Sync Online";
            document.getElementById('timezone-label').innerText = Intl.DateTimeFormat().resolvedOptions().timeZone;
            serverError = false;
        }

        let tableHtml = '';
        if (data.ntp_response) {
            tableHtml += `<tr><td class="label">NTP Server: </td><td class="value">${data.ntp_response['server']}</td></tr>`;
        }
        tableHtml += `<tr><td class="label">DateTime (ISO): </td><td class="value">${data.datetime}</td></tr>`;
        tableHtml += `<tr><td class="label">Timestamp (UNIX): </td><td class="value">${data.timestamp}</td></tr>`;
        if (data.ntp_response) {
            tableHtml += `<tr><td class="label">Stratum: </td><td class="value">${data.ntp_response['stratum']}</td></tr>`;
            tableHtml += `<tr><td class="label">Root Dispersion: </td><td class="value">${data.ntp_response['root_dispersion']}</td></tr>`;
            tableHtml += `<tr><td class="label">Root Distance: </td><td class="value">${data.ntp_response['root_distance']}</td></tr>`;
            tableHtml += `<tr><td class="label">Round-Trip Time: </td><td class="value">${data.ntp_response['rtt']}</td></tr>`;

        }
        document.getElementById('data-table').innerHTML = tableHtml;

    } catch (e) {
        document.getElementById('main-card').classList.add('error-state');
        document.getElementById('turkey-icon').src = "https://aecrypto.io/static/cartoons/turkey_depressed.png";
        document.getElementById('clock').innerText = "OFFLINE";
        serverError = true;
    }
}

async function syncWatch() {
    await syncWithNTP();
    for (let i = 1; i <= 16; i++) {
        await sleep(1000);   
        currentTimestamp += 1000;
        formatLocal();
    }
}

async function main() {
    formatLocal();
    while (true) {
        await syncWatch();
    }
}
main()
