/* eslint-disable react/destructuring-assignment */
import React from 'react';
import { useDispatch } from 'react-redux';
import IntegrationTableView from '../../common/integration-table-view/integration-table-view';
import AppLoader from '../../common/app-loader/app-loader';
import {
  requestIntegrationDelete,
  showModal,
  resetIntegrationStates,
} from '../../../actions/app-actions';
import { NO_INTEGRATION_FOUND_ALERT_MESSAGE } from '../../../constants/visualization-config';

function getEmptyStateView() {
  return (
    <div className="empty-state-wrapper">
      {NO_INTEGRATION_FOUND_ALERT_MESSAGE.message}
    </div>
  );
}

function getTableEmptyState(data) {
  const emptyStateWrapper = {
    height: '400px',
    width: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  };
  return (
    <div style={emptyStateWrapper}>
      {data === undefined ? <AppLoader /> : getEmptyStateView()}
    </div>
  );
}

function isDataAvailable(data) {
  let result;
  if (data && data.length > 0) {
    result = true;
  } else {
    result = false;
  }
  return result;
}

const HTTPEndpointList = props => {
  const dispatch = useDispatch();

  const resetStates = () => {
    dispatch(resetIntegrationStates());
  };

  const getIntegrationTableView = () => {
    const { httpEndpointList } = props;
    return (
      <IntegrationTableView
        recordCollection={httpEndpointList}
        onDeleteRequestCallback={record => handleDeleteDialog(record)}
      />
    );
  };

  const deleteIntegration = record => {
    const params = {
      id: record.id,
      notification_type: record.notification_type,
    };
    return dispatch(requestIntegrationDelete(params));
  };

  const handleDeleteDialog = record => {
    const params = {
      dialogTitle: 'Delete Integration?',
      dialogBody:
        'Are you sure you want to delete this HTTP enpoint integration?',
      confirmButtonText: 'Yes, Delete',
      cancelButtonText: 'No, Keep',
      onConfirmButtonClick: () => deleteIntegration(record),
    };
    dispatch(showModal('DIALOG_MODAL', params));
    resetStates();
  };

  const { httpEndpointList } = props;
  return (
    <div className="integration-list-section">
      {isDataAvailable(httpEndpointList)
        ? getIntegrationTableView()
        : getTableEmptyState(httpEndpointList)}
    </div>
  );
};

export default HTTPEndpointList;
