const tsPreset = require('ts-jest/jest-preset');

module.exports = {
  ...tsPreset,
  globals: {
    test_url: `http://${process.env.HOST || 'localhost'}:${process.env.PORT || 8080}`,
  },
};
