import type { ConfigFile } from '@rtk-query/codegen-openapi';

const config: ConfigFile = {
  schemaFile: 'public/openapi3.json',
  apiFile: '', // leave this empty, and instead populate the outputFiles object below
  hooks: true,
  tag: true,

  outputFiles: {
    './public/app/features/migrate-to-cloud/api/endpoints.gen.ts': {
      apiFile: './public/app/features/migrate-to-cloud/api/baseAPI.ts',
      apiImport: 'baseAPI',
      filterEndpoints: [
        'getFolders',
        'getFolderByUid',
        'createFolder',
        'deleteFolder',
        'updateFolder',
        'getFolderDescendantCounts',
      ],
    },
  },
};

export default config;
