module.exports = {
  preset: './jest-earth-rest-api-preset.js',
  testSequencer: './jest-test-sequencer.js',
  roots: ['run'],
  verbose: false,
  collectCoverage: false,
  testTimeout: 70 * 1000,
  reporters: [
    'default',
    [
      './node_modules/jest-html-reporter',
      {
        pageTitle: 'earth-rest-api',
        dateFormat: 'dd.mm.yyyy HH:MM:ss o',
        // outputPath: '/tms-test/output/test-report.html', // <test.DirOutput>/test-report.html
        includeFailureMsg: true,
        includeSuiteFailure: true,
      },
    ],
  ],
};
