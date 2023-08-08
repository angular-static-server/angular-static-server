import 'angular-server-side-configuration/process';

/**
 * How to use angular-server-side-configuration:
 *
 * Use process.env['NAME_OF_YOUR_ENVIRONMENT_VARIABLE']
 *
 * const stringValue = process.env['STRING_VALUE'];
 * const stringValueWithDefault = process.env['STRING_VALUE'] || 'defaultValue';
 * const numberValue = Number(process.env['NUMBER_VALUE']);
 * const numberValueWithDefault = Number(process.env['NUMBER_VALUE'] || 10);
 * const booleanValue = process.env['BOOLEAN_VALUE'] === 'true';
 * const booleanValueInverted = process.env['BOOLEAN_VALUE_INVERTED'] !== 'false';
 * const complexValue = JSON.parse(process.env['COMPLEX_JSON_VALUE]);
 * 
 * Please note that process.env[variable] cannot be resolved. Please directly use strings.
 */


export const environment = {
  label: process.env['LABEL'],
  cspNonce: process.env['NGSSC_CSP_NONCE'],
};
