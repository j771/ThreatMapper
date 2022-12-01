import React from 'react';
import { connect } from 'react-redux';
import DFSelect from '../../common/multi-select/app';
import IntegrationTableView from '../../common/integration-table-view/integration-table-view';

import {
  requestIntegrationDelete,
  resetIntegrationStates,
  submitIntegrationRequest,
  showModal,
  enumerateFiltersAction,
} from '../../../actions/app-actions';
import { getIntegrations } from '../../../utils/web-api-utils';
import AppLoader from '../../common/app-loader/app-loader';

import {
  FEATURE_BLOCKED_ALERT_MESSAGE,
  NO_INTEGRATION_FOUND_ALERT_MESSAGE,
} from '../../../constants/visualization-config';

const resourceCollection = [
  {
    name: 'Vulnerabilities',
    value: 'vulnerability',
  }
];

const allNodeType = 'host,container_image,pod,aws';

const isDataAvailable = (data) => {
  let result;
  if (data && data.length > 0) {
    result = true;
  } else {
    result = false;
  }
  return result;
}
const getDisabledBtnView = () => {
  return (
    <button type="button" className="app-btn">
      Subscribe
    </button>
  );
}

const getEnabledBtnView = (handleSubmit) => {
  return (
    <button type="button" className="app-btn" onClick={handleSubmit}>
      Subscribe
    </button>
  );
}

class AWSSecurityLakeIntegrationView extends React.Component {
  constructor() {
    super();
    this.state = {
      isSuccess: false,
      isError: false,
      submitted: false,
      filters: {},
      cloudTrailValue: {},
      awsSecurityAccountFilter: {},
      awsAccessKey: '',
      awsSecretKey: '',
      awsRegion: '',
      resourceType: '',
      role: '',
    };
    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
    this.handleResourceChange = this.handleResourceChange.bind(this);
    this.deleteIntegration = this.deleteIntegration.bind(this);
    this.handleDeleteDialog = this.handleDeleteDialog.bind(this);
    this.seCloudtrailOptions = this.seCloudtrailOptions.bind(this);
  }

  componentDidMount() {
    this.resetStates();
    this.fetchIntegrationList();
    this.getFilters();
  }

  getFilters() {
    const { dispatch } = this.props;

    const params = {
      node_type: allNodeType,
      filters:
        'host_name,container_name,image_name_with_tag,user_defined_tags,kubernetes_namespace,kubernetes_cluster_name,cloudtrail_trail,aws_account_id',
    };
    return dispatch(enumerateFiltersAction(params));
  }

  UNSAFE_componentWillReceiveProps(newProps) {
    if (newProps.isSuccess && !newProps.isError) {
      this.setState({
        integrationAddResponse: newProps.integrationAddResponse,
        isSuccess: newProps.isSuccess,
        isError: newProps.isError,
      });
    } else if (!newProps.isSuccess && newProps.isError) {
      this.setState({
        integrationAddResponse: newProps.integrationAddResponse,
        isSuccess: newProps.isSuccess,
        isError: newProps.isError,
      });
    }

    const { integrationAddResponse: currIntegrationAddResponse } = this.props;
    if (newProps.integrationAddResponse !== currIntegrationAddResponse) {
      this.setState({
        integrationAddResponse: newProps.integrationAddResponse,
      });
    }

    if (newProps.licenseResponse) {
      this.setState({
        isDemoModeEnabled: newProps.licenseResponse.demo_mode,
      });
    }
  }

  resetStates() {
    this.props.dispatch(resetIntegrationStates());
  }

  fetchIntegrationList() {
    getIntegrations(this.props.dispatch);
  }

  handleChange(e) {
    const { name, value } = e.target;
    this.setState({ [name]: value });
  }

  handleResourceChange(e) {
    const selectedOption = e;
    this.setState({
      resourceType: selectedOption,
    });
  }

  handleSubmit(e) {
    e.preventDefault();
    this.setState({ submitted: true });
    const {
      resourceType,
      filters,
      cloudTrailValue,
      awsSecurityAccountFilter,
      role
    } = this.state;

    if (!resourceType) {
      this.setState({
        isError: true,
      });
      return;
    }

    if (resourceType && resourceType.value === 'cloudtrail_alerts') {
      if (Object.keys(cloudTrailValue).length === 0) {
        this.setState({
          integrationAddResponse: 'CloudTrail selection is mandatory',
          isError: true,
        });
        return;
      }
    }

    this.setState({
      integrationAddResponse: '',
      isError: false,
    });

    const apiFilters = Object.keys(filters).reduce((acc, key) => {
      acc[key] = filters[key].map(el => el.value);
      return acc;
    }, {});

    const apiCloudTrailFilters = Object.keys(cloudTrailValue).reduce(
      (acc, key) => {
        acc[key] = cloudTrailValue[key].map(el => el.value);
        return acc;
      },
      {}
    );

    const apiAWSAccountFilters = Object.keys(awsSecurityAccountFilter).reduce(
      (acc, key) => {
        acc[key] = awsSecurityAccountFilter[key].map(el => el.value);
        return acc;
      },
      {}
    );

    const filterObject = {
      ...apiFilters,
      ...apiCloudTrailFilters,
      ...apiAWSAccountFilters,
    };

    const params = {
      aws_access_key: this.state.awsAccessKey,
      aws_secret_key: this.state.awsSecretKey,
      region_name: this.state.awsRegion,
      integration_type: 'amazon_security_lake',
      notification_type: resourceType.value,
      glue_iam_role_arn: role,
      filters: filterObject,
    };

    this.props.dispatch(submitIntegrationRequest(params));
  }

  seCloudtrailOptions(name, value) {
    this.setState({
      cloudTrailValue: {
        // eslint-disable-next-line
        ...this.state.cloudTrailValue,
        cloudtrail_trail: value,
      },
    });
  }

  setAwsSecurityOptions(name, value) {
    this.setState({
      awsSecurityAccountFilter: {
        // eslint-disable-next-line
        ...this.state.awsSecurityAccountFilter,
        aws_account_id: value,
      },
    });
  }

  getIntegrationFormView() {
    const {
      submitted,
      isDemoModeEnabled,
      resourceType,
      cloudTrailValue,
      awsSecurityAccountFilter,
    } = this.state;
    const {
      awsAccessKey,
      awsSecretKey,
      awsRegion,
      role,
    } = this.state;
    const cloudTrailOptions =
      this.props.nodeFilters &&
      this.props.nodeFilters.filter((item) => {
        return item.label === 'CloudTrail';
      });
    const columnStyle = {
      padding: '0px 60px',
    };
    // creating accounts options
    const AWSAccountOptions =
      this.props.nodeFilters &&
      this.props.nodeFilters.filter(item => {
        return item.name === 'aws_account_id';
      });

    return (
      <div className="form-wrapper">
        <form name="form">
          <div className="" style={{ columnStyle }}>
            <div className="row">
              <div className="col-md-4">
                <div
                  className={`form-group ${
                    submitted && !awsAccessKey ? 'has-error' : ''
                  }`}
                >
                  <label htmlFor="awsAccessKey">
                    <i className="fa fa-key" aria-hidden="true" />
                    <input
                      type="text"
                      className="form-control"
                      name="awsAccessKey"
                      placeholder="AWS Access Key"
                      value={awsAccessKey}
                      onChange={this.handleChange}
                      autoComplete="off"
                    />
                  </label>
                  {submitted && !awsAccessKey && (
                    <div className="field-error">
                      AWS Access Key is required
                    </div>
                  )}
                </div>
              </div>
              <div className="col-md-4">
                <div
                  className={`form-group ${
                    submitted && !awsSecretKey ? 'has-error' : ''
                  }`}
                >
                  <label htmlFor="awsSecretKey">
                    <i className="fa fa-key" aria-hidden="true" />
                    <input
                      type="text"
                      className="form-control"
                      name="awsSecretKey"
                      placeholder="AWS Secret Key"
                      value={awsSecretKey}
                      onChange={this.handleChange}
                      autoComplete="off"
                    />
                  </label>
                  {submitted && !awsSecretKey && (
                    <div className="field-error">
                      AWS Secret Key is required
                    </div>
                  )}
                </div>
              </div>
            </div>
            <div className="row">
              <div className="col-md-4">
                <div
                  className={`form-group ${
                    submitted && !awsRegion ? 'has-error' : ''
                  }`}
                >
                  <label htmlFor="awsRegion">
                    <i className="fa fa-globe" aria-hidden="true" />
                    <input
                      type="text"
                      className="form-control"
                      name="awsRegion"
                      placeholder="AWS Region"
                      value={awsRegion}
                      onChange={this.handleChange}
                      autoComplete="off"
                    />
                  </label>
                  {submitted && !awsRegion && (
                    <div className="field-error">AWS Region is required</div>
                  )}
                </div>
              </div>
            </div>

            <div className="row">
              <div className="col-md-4">
                <div className="form-group df-select-field">
                  {AWSAccountOptions?.map((resource, index) => (
                    <div className="search-form" key={resource.value}>
                      <DFSelect
                        options={resource.options.map(el => ({
                          label: el,
                          value: el,
                        }))}
                        name={resource.name}
                        placeholder={`${resource.label}`}
                        onChange={selectedOptions =>
                          this.setAwsSecurityOptions(
                            resource.name,
                            selectedOptions
                            )
                          }
                          value={cloudTrailValue[resource.name]}
                          />
                      {index === 0 && <div style={{ fontSize: '12px', paddingTop: '8px', paddingLeft: '8px' }}>
                        Amazon Security Lake must be enabled for this account and the given credentials should have access to write data to the given account&apos;s data lake.
                      </div>}
                    </div>
                  ))}
                  {submitted &&
                    (!awsSecurityAccountFilter.aws_account_id ||
                      awsSecurityAccountFilter?.aws_account_id?.length ===
                        0) && (
                      <div
                        className="field-error"
                        style={{
                          marginTop: '10px',
                        }}
                      >
                        AWS Acccount selection is required
                      </div>
                    )}
                </div>
              </div>
              <div className="col-md-4">
                <div
                  className={`form-group ${
                    submitted && !role ? 'has-error' : ''
                  }`}
                >
                  <label htmlFor="role">
                    <i className="fa fa-user-o" aria-hidden="true" />
                    <input
                      type="text"
                      className="form-control"
                      name="role"
                      placeholder="Glue IAM Role ARN"
                      value={role}
                      onChange={this.handleChange}
                      autoComplete="off"
                    />
                  </label>
                  <div style={{ fontSize: '12px', paddingLeft: '8px' }}>{
                  `Provide a role that Security Lake will use to run glue crawlers in your accounts. 
                  This role must have the AWSGlueServiceRole managed policy attached to it. 
                  It must allow full access to your data lake and the trust policy must include the Glue service principal. 
                  If you have used your own encryption keys, this role must allow access to kms:Decrypt.`}</div>
                  {submitted && !role && (
                    <div className="field-error">Glue IAM Role ARN is required</div>
                  )}
                </div>
              </div>
            </div>

            <div className="row">
              <div className="col-md-4">
                  <div
                    className="severity-option-wrapper"
                    onChange={this.handleResourceChange}
                  >
                    <div className="wrapper-heading">Choose Resource*</div>
                    <div className="resource-option-wrapper">
                      <div className="form-group df-select-field">
                        <DFSelect
                          options={resourceCollection.map(el => ({
                            value: el.value,
                            label: el.name,
                          }))}
                          onChange={this.handleResourceChange}
                          placeholder="Resources"
                          value={this.state.resourceType}
                          clearable={false}
                        />
                        {submitted && !resourceType && (
                          <div
                            className="field-error"
                            style={{
                              marginTop: '10px',
                            }}
                          >
                            Resource selection is required
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

            {resourceType && resourceType.value === 'cloudtrail_alert' && (
              <div className="row">
                <div className="col">
                  <div
                    className="form-group df-select-field"
                    style={{ width: '250px' }}
                  >
                    {cloudTrailOptions.map(filter => (
                      <div className="search-form" key={filter.value}>
                        <br />
                        <DFSelect
                          options={filter.options.map(el => ({
                            label: el,
                            value: el,
                          }))}
                          name={filter.name}
                          placeholder={`${filter.label}`}
                          onChange={selectedOptions =>
                            this.seCloudtrailOptions(
                              filter.name,
                              selectedOptions
                            )
                          }
                          value={cloudTrailValue[filter.name]}
                          isMulti
                        />
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}
            <div className="error-msg-container">
              {this.state.isError && (
                <div className="auth-error-msg">
                  {this.state.integrationAddResponse}
                </div>
              )}
            </div>
            <br />
            <div className="form-group">
              {isDemoModeEnabled
                ? getDisabledBtnView()
                : getEnabledBtnView(this.handleSubmit)}
            </div>
            <div className="error-msg-container">
              {this.state.isSuccess && (
                <div className="auth-success-msg">
                  {this.state.integrationAddResponse}
                </div>
              )}
            </div>
          </div>
        </form>
      </div>
    );
  }

  deleteIntegration(record) {
    const params = {
      id: record.id,
      notification_type: record.notification_type,
    };
    return this.props.dispatch(requestIntegrationDelete(params));
  }

  handleDeleteDialog(record) {
    const params = {
      dialogTitle: 'Delete Integration?',
      dialogBody:
        'Are you sure you want to delete the AWS security hub integration?',
      confirmButtonText: 'Yes, Delete',
      cancelButtonText: 'No, Keep',
      onConfirmButtonClick: () => this.deleteIntegration(record),
    };
    this.props.dispatch(showModal('DIALOG_MODAL', params));
    this.resetStates();
  }

  getIntegrationTableView() {
    const results = this.props.availableAWSSecurityLakeIntegrations.reduce((acc,data) => {
      acc = [...acc, {
        ...data,
        filters: data.filters.aws_account_id,
      }]
      return acc;
    }, [])
    return (
      <IntegrationTableView
        recordCollection={results}
        onDeleteRequestCallback={record => this.handleDeleteDialog(record)}
      />
    );
  }

  getTableEmptyState(data) {
    const emptyStateWrapper = {
      height: '400px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
    };
    return (
      <div style={emptyStateWrapper}>
        {data === undefined ? <AppLoader /> : this.getEmptyStateView()}
      </div>
    );
  }

  getEmptyStateView() {
    const { isDemoModeEnabled } = this.state;
    return (
      <div className="empty-state-wrapper">
        {isDemoModeEnabled
          ? FEATURE_BLOCKED_ALERT_MESSAGE.message
          : NO_INTEGRATION_FOUND_ALERT_MESSAGE.message}
      </div>
    );
  }

  render() {
    const { availableAWSSecurityLakeIntegrations } = this.props;
    return (
      <div className="email-integration-view-wrapper">
        <div className="integration-form-section">
          {this.getIntegrationFormView()}
        </div>
        <div className="integration-list-section">
          {isDataAvailable(availableAWSSecurityLakeIntegrations)
            ? this.getIntegrationTableView()
            : this.getTableEmptyState(availableAWSSecurityLakeIntegrations)}
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    isSuccess: state.get('isSuccess'),
    isError: state.get('isError'),
    integrationAddResponse: state.get('integrationAddResponse'),
    availableAWSSecurityLakeIntegrations: state.get('availableAWSSecurityLakeIntegrations'),
    licenseResponse: state.get('licenseResponse'),
    nodeFilters: state.getIn(['nodesView', 'topologyFilters', allNodeType]),
  };
}

export default connect(mapStateToProps)(AWSSecurityLakeIntegrationView);