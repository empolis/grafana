import React, { PureComponent } from 'react';
import { FieldSet, Field, RadioButtonGroup, Modal, Button } from '@grafana/ui';
import { SelectableValue, PanelModel, AppEvents } from '@grafana/data';
import { DashboardModel } from 'app/features/dashboard/state';
import { appEvents } from 'app/core/core';
import { buildPdfUrl } from './utils';

const themeOptions: Array<SelectableValue<string>> = [
  { label: 'Current', value: 'current' },
  { label: 'Dark', value: 'dark' },
  { label: 'Light', value: 'light' },
];

const pdfOptions: Array<SelectableValue<string>> = [
  { label: 'Portrait', value: 'portrait' },
  { label: 'Landscape', value: 'landscape' },
];

export interface Props {
  dashboard: DashboardModel;
  panel?: PanelModel;
  onDismiss(): void;
}

export interface State {
  selectedTheme: string;
  pdfOrientation: string;
}

export class SharePdf extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      selectedTheme: 'current',
      pdfOrientation: 'portrait',
    };
  }

  onThemeChange = (value: string) => {
    this.setState({ selectedTheme: value });
  };

  onOrientationChange = (value: string) => {
    this.setState({ pdfOrientation: value });
  };

  generatePdf = () => {
    const { panel } = this.props;
    const { selectedTheme, pdfOrientation } = this.state;

    window.open(buildPdfUrl(pdfOrientation === 'landscape', selectedTheme, panel), '_blank');
    appEvents.emit(AppEvents.alertSuccess, ['PDF opened in new tab']);
    this.props.onDismiss();
  };

  render() {
    const { onDismiss } = this.props;
    const { selectedTheme, pdfOrientation } = this.state;

    return (
      <>
        <p className="share-modal-info-text">
          Render dashboard as a PDF document.
        </p>
        <FieldSet>
          <Field label="Orientation">
            <RadioButtonGroup options={pdfOptions} value={pdfOrientation} onChange={this.onOrientationChange} />
          </Field>
          <Field label="Theme">
            <RadioButtonGroup options={themeOptions} value={selectedTheme} onChange={this.onThemeChange} />
          </Field>
        </FieldSet>
        <Modal.ButtonRow>
          <Button variant="secondary" onClick={onDismiss} fill="outline">
            Cancel
          </Button>
          <Button variant="primary" onClick={this.generatePdf}>
            Save as PDF
          </Button>
        </Modal.ButtonRow>
      </>
    );
  }
}
