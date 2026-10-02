/**
 * Personal Finance Analytics Google Sheets integration.
 *
 * Configure once from the Apps Script editor:
 *   setGatewayConfig('http://localhost:8080', '<jwt-access-token>');
 *
 * The token is stored in UserProperties and is never written to a sheet.
 */

const CONFIG_KEYS = {
  baseURL: 'PFA_GATEWAY_BASE_URL',
  accessToken: 'PFA_ACCESS_TOKEN',
};

function onOpen() {
  SpreadsheetApp.getUi()
    .createMenu('Personal Finance')
    .addItem('Create transaction from active row', 'createTransactionFromActiveRow')
    .addItem('Get report for current month', 'writeCurrentMonthSummary')
    .addToUi();
}

function setGatewayConfig(baseURL, accessToken) {
  if (!baseURL || !accessToken) {
    throw new Error('baseURL and accessToken are required');
  }

  PropertiesService.getUserProperties().setProperties({
    [CONFIG_KEYS.baseURL]: baseURL.replace(/\/$/, ''),
    [CONFIG_KEYS.accessToken]: accessToken,
  });
}

function createTransaction(amount, category, description, occurredAt) {
  const payload = {
    amount: Number(amount),
    category: String(category),
    description: String(description || ''),
    occurred_at: new Date(occurredAt).toISOString(),
  };

  return requestGateway_('/transactions', 'post', payload);
}

function createTransactionFromActiveRow() {
  const row = SpreadsheetApp.getActiveSheet().getActiveRange().getValues()[0];
  if (row.length < 4) {
    throw new Error('Select a row with Date, Amount, Category and Description columns');
  }

  const result = createTransaction(row[1], row[2], row[3], row[0]);
  SpreadsheetApp.getActiveSheet().getRange('F1').setValue('Last transaction ID');
  SpreadsheetApp.getActiveSheet().getRange('F2').setValue(result.id);
}

function getSummary(from, to) {
  const query = '?from=' + encodeURIComponent(new Date(from).toISOString()) +
    '&to=' + encodeURIComponent(new Date(to).toISOString());

  return requestGateway_('/reports/summary' + query, 'get');
}

function writeCurrentMonthSummary() {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1);
  const to = new Date(now.getFullYear(), now.getMonth() + 1, 1);
  const summary = getSummary(from, to);
  const sheet = SpreadsheetApp.getActiveSheet();

  sheet.getRange('H1:I4').clearContent();
  sheet.getRange('H1:I4').setValues([
    ['Metric', 'Value'],
    ['From', summary.from],
    ['To', summary.to],
    ['Total spent', summary.total_spent],
  ]);
}

function requestGateway_(path, method, payload) {
  const properties = PropertiesService.getUserProperties();
  const baseURL = properties.getProperty(CONFIG_KEYS.baseURL);
  const accessToken = properties.getProperty(CONFIG_KEYS.accessToken);
  if (!baseURL || !accessToken) {
    throw new Error('Run setGatewayConfig(baseURL, accessToken) first');
  }

  const options = {
    method: method,
    headers: {Authorization: 'Bearer ' + accessToken},
    muteHttpExceptions: true,
  };
  if (payload !== undefined) {
    options.contentType = 'application/json';
    options.payload = JSON.stringify(payload);
  }

  const response = UrlFetchApp.fetch(baseURL + path, options);
  const status = response.getResponseCode();
  const body = response.getContentText();
  if (status < 200 || status >= 300) {
    throw new Error('Gateway request failed (' + status + '): ' + body);
  }

  return JSON.parse(body);
}
