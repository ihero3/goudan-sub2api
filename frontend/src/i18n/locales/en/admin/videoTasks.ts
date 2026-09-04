export default {
  admin: {
    videoTasks: {
      title: 'Video Tasks',
      description: 'Manage video generation tasks across multiple upstream models (MiniMax-H3 / Seedance / Wan3.0-Video).',
      userIdPlaceholder: 'Enter user ID to query',
      emptyStateTitle: 'Enter a user ID',
      userIdRequired: 'Enter a user ID above to start querying that user\'s video tasks.',
      noData: 'No video tasks',
      tryOtherFilters: 'No video tasks match the current filters. Try adjusting them.',
      loadFailed: 'Failed to load video tasks',
      cancelSuccess: 'Task cancelled',
      cancelFailed: 'Failed to cancel task',
      cancel: 'Cancel',
      cancelConfirmTitle: 'Cancel task',
      cancelConfirmMessage: 'Are you sure you want to cancel task #{id}? It will be marked as cancelled.',
      openVideo: 'Open',
      finishedAt: 'Finished',
      status: {
        processing: 'Processing',
        succeeded: 'Succeeded',
        failed: 'Failed',
        cancelled: 'Cancelled'
      },
      columns: {
        taskId: 'Task ID',
        model: 'Model',
        status: 'Status',
        user: 'User',
        channel: 'Channel',
        resolution: 'Spec',
        video: 'Video',
        cost: 'Cost',
        error: 'Error',
        createdAt: 'Created At'
      }
    }
  }
}
