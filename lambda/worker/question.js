exports.handler = async (event) => {
  console.log(
    JSON.stringify({
      appEnvironment: process.env.APP_ENV,
      databaseHost: process.env.DATABASE_HOST,
      lambdaName: process.env.LAMBDA_NAME,
      recordCount: event.Records?.length ?? 0,
    }),
  );
};
