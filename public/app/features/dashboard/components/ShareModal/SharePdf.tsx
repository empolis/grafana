import React, { PureComponent } from 'react';
import { LegacyForms, Icon, Button } from '@grafana/ui';
const { Select } = LegacyForms;
import { SelectableValue, PanelModel } from '@grafana/data';
import { DashboardModel } from 'app/features/dashboard/state';
import { buildPdfUrl } from './utils';

const themeOptions: Array<SelectableValue<string>> = [
  { label: 'current', value: 'current' },
  { label: 'dark', value: 'dark' },
  { label: 'light', value: 'light' },
];

export interface Props {
  dashboard: DashboardModel;
  panel?: PanelModel;
  empolisShare?: boolean;
}

export interface State {
  selectedTheme: SelectableValue<string>;
}

export class SharePdf extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      selectedTheme: themeOptions[0],
    };
  }

  onThemeChange = (value: SelectableValue<string>) => {
    this.setState({ selectedTheme: value });
  };

  generatePdf = () => {
    const { panel } = this.props;
    const { selectedTheme } = this.state;

    window.open(buildPdfUrl(false, selectedTheme.value, panel), '_blank');
  };

  generateLandscapePdf = () => {
    const { panel } = this.props;
    const { selectedTheme } = this.state;

    window.open(buildPdfUrl(true, selectedTheme.value, panel), '_blank');
  };

  render() {
    const { selectedTheme } = this.state;

    return (
      <div className="share-modal-body">
        <div className="share-modal-header">
          <Icon name="file-alt" className="share-modal-big-icon" size="xxl" />
          <div className="share-modal-content">
            <div>
              <div className="gf-form">
                <label className="gf-form-label width-12">Theme</label>
                <Select width={10} options={themeOptions} value={selectedTheme} onChange={this.onThemeChange} />
              </div>
              <div className="gf-form">
                <div className="gf-form-row">
                  <Button className="btn width-16" variant="primary" onClick={this.generatePdf}>
                    <Icon name="file-alt" /> Generate Portrait PDF
                  </Button>
                  <Button className="btn width-16" variant="secondary" onClick={this.generateLandscapePdf}>
                    <Icon name="file-alt" /> Generate Landscape PDF
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }
}
