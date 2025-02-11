const fs = require('fs');

const exportData = JSON.parse(fs.readFileSync('../export.json', 'utf8'));
const importData = {
  PloEquity: exportData.Items.map(item => ({
    PutRequest: {
      Item: item
    }
  }))
};

fs.writeFileSync('../import-items.json', JSON.stringify(importData, null, 2));
console.log('Conversion completed: import-items.json has been created.');
