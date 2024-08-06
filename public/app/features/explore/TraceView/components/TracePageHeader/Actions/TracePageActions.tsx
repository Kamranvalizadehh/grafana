import { css } from '@emotion/css';
import { useState } from 'react';

import { GrafanaTheme2, CoreApp, DataFrame } from '@grafana/data';
import { reportInteraction } from '@grafana/runtime';
import { Button, Dropdown, Icon, Menu, Modal, useTheme2 } from '@grafana/ui';

import { config } from '../../../../../../core/config';
import { downloadTraceAsJson, exportTraceAsMermaid } from '../../../../../inspector/utils/download';

import ActionButton from './ActionButton';

export const getStyles = (theme: GrafanaTheme2) => {
  return {
    TracePageActions: css({
      label: 'TracePageActions',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      gap: '4px',
      marginBottom: '10px',
    }),
    feedback: css({
      margin: '6px 6px 6px 0',
      color: theme.colors.text.secondary,
      fontSize: theme.typography.bodySmall.fontSize,
      '&:hover': {
        color: theme.colors.text.link,
      },
    }),
  };
};

export type TracePageActionsProps = {
  traceId: string;
  data: DataFrame;
  app?: CoreApp;
};

export default function TracePageActions(props: TracePageActionsProps) {
  const { traceId, data, app } = props;
  const theme = useTheme2();
  const styles = getStyles(theme);
  const [copyTraceIdClicked, setCopyTraceIdClicked] = useState(false);
  const [exportModalOpen, setExportModalOpen] = useState(false);
  const [mermaidDiagramCode, setMermaidDiagramCode] = useState('');

  const copyTraceId = () => {
    navigator.clipboard.writeText(traceId);
    setCopyTraceIdClicked(true);
    setTimeout(() => {
      setCopyTraceIdClicked(false);
    }, 5000);
  };

  const exportTrace = () => {
    const traceFormat = downloadTraceAsJson(data, 'Trace-' + traceId.substring(traceId.length - 6));
    reportInteraction('grafana_traces_download_traces_clicked', {
      app,
      grafana_version: config.buildInfo.version,
      trace_format: traceFormat,
      location: 'trace-view',
    });
  };

  const exportMermaid = () => {
    setMermaidDiagramCode(exportTraceAsMermaid(data, traceId));
    setExportModalOpen(true);
  };
  const exportMenu = (
    <Menu>
      <Menu.Item label="Native Download" onClick={exportTrace} />
      <Menu.Item label="Mermaid" onClick={exportMermaid} />
    </Menu>
  );

  return (
    <div className={styles.TracePageActions}>
      <a
        href="https://forms.gle/RZDEx8ScyZNguDoC8"
        className={styles.feedback}
        title="Share your thoughts about tracing in Grafana."
        target="_blank"
        rel="noreferrer noopener"
      >
        <Icon name="comment-alt-message" /> Give feedback
      </a>

      <ActionButton
        onClick={copyTraceId}
        ariaLabel={'Copy Trace ID'}
        label={copyTraceIdClicked ? 'Copied!' : 'Trace ID'}
        icon={'copy'}
      />
      <Dropdown overlay={exportMenu}>
        <Button size="sm" variant="secondary" fill={'outline'} type="button" icon={'save'}>
          Export
        </Button>
      </Dropdown>
      <ActionButton onClick={exportMermaid} ariaLabel={'Export Mermaid'} label={'Export Mermaid'} icon={'chart-line'} />
      <Modal isOpen={exportModalOpen} onDismiss={() => setExportModalOpen(false)} title={'Mermaid Export'}>
        <pre>{mermaidDiagramCode}</pre>
      </Modal>
    </div>
  );
}
