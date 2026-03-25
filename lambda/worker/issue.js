exports.handler = async (event) => {
  console.log(
    JSON.stringify({
      appEnvironment: process.env.APP_ENV,
      databaseHost: process.env.DATABASE_HOST,
      issueApiHostname: process.env.ISSUE_API_HOSTNAME,
      lambdaName: process.env.LAMBDA_NAME,
      recordCount: event.Records?.length ?? 0,
    }),
  );
};
