/* eslint-disable */
const path = require('path');
const fs = require('fs');

const HtmlWebPackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');

// ====== CONFIG ======
const buildDir = path.resolve(__dirname, 'dist');

function getFirstPartFileName(s) {
  return s.split('.')[0];
}

// Buildable “modules”
const entries = {
  'infrastructure-manage': { 'infrastructure-manage': './index.js' },
};

// Html plugins per module
const htmlPlugins = {
  'infrastructure-manage': [
    new HtmlWebPackPlugin({
      // We manually inject via template tags below
      inject: 'body',
      chunks: ['infrastructure-manage'],
      // NEW: output here
      filename: path.resolve(__dirname, '../../../views/infrastructure-manage.html'),
      // Template here
      template: path.resolve(__dirname, 'template/infrastructure-manage.html'),
    }),
  ],
};

// Webpack 5-compatible plugin to prune old hashed assets sharing the same basename
class PruneOldAssetsPlugin {
  constructor(options = {}) {
    this.buildDir = options.buildDir;
    this.getFirstPart = options.getFirstPart || ((s) => s);
  }
  apply(compiler) {
    compiler.hooks.afterEmit.tap('PruneOldAssetsPlugin', (compilation) => {
      if (!this.buildDir) return;

      const newlyCreatedAssets = {};
      for (const asset of compilation.getAssets()) {
        newlyCreatedAssets[asset.name] = true;
      }
      const newBaseNames = Object.keys(newlyCreatedAssets).map(this.getFirstPart);

      fs.readdir(this.buildDir, (err, files = []) => {
        if (err) return; // dist may not exist initially
        const unlinked = [];
        files.forEach((file) => {
          const base = this.getFirstPart(file);
          if (!newlyCreatedAssets[file] && newBaseNames.includes(base)) {
            try {
              fs.unlinkSync(path.join(this.buildDir, file));
              unlinked.push(file);
            } catch (_) {}
          }
        });
        if (unlinked.length) console.log('Removed old assets:', unlinked);
      });
    });
  }
}

module.exports = (w, env = {}) => {
  const mod = env.module || env.buildModule || 'infrastructure-manage';
  if (!entries[mod]) {
    throw new Error(`Unknown module "${mod}". Available: ${Object.keys(entries).join(', ')}`);
  }

  const isProd =
      env.mode === 'production' ||
      process.env.NODE_ENV === 'production' ||
      (typeof w === 'object' && w.mode === 'production');

  const plugins = [
    ...(htmlPlugins[mod] || []),
    new MiniCssExtractPlugin({
      filename: '[name].[fullhash].css', // versioned CSS
    }),
    new PruneOldAssetsPlugin({
      buildDir,
      getFirstPart: getFirstPartFileName,
    }),
  ];

  return {
    mode: isProd ? 'production' : 'development',

    entry: entries[mod],

    output: {
      path: buildDir,
      filename: '[name].bundle.[fullhash].js', // versioned JS
      chunkFilename: '[name].[fullhash].chunk.js',
      publicPath: '/static/js/webpack/dist/', // used by injected script/link tags
      clean: false,
    },

    module: {
      rules: [
        {
          test: /\.jsx?$/,
          exclude: /node_modules\/(?!(highcharts)\/).*/,
          use: {
            loader: 'babel-loader',
            options: {
              presets: [
                ['@babel/preset-env', { targets: 'defaults' }],
                ['@babel/preset-react', { runtime: 'automatic' }],
              ],
            },
          },
        },
        {
          test: /\.css$/,
          exclude: /node_modules/,
          use: [MiniCssExtractPlugin.loader, 'css-loader'],
        },
        {
          test: /\.s[ac]ss$/i,
          use: [MiniCssExtractPlugin.loader,
            'css-loader',
            'sass-loader'],
        },
      ],
    },

    optimization: {
      minimize: isProd,
      minimizer: [
        '...', // default Terser
        new CssMinimizerPlugin(),
      ],
      // splitChunks: { chunks: 'all' },
    },

    // Do NOT externalize React unless you ONLY use CDN UMD builds.
    // externals: { react: 'React', 'react-dom': 'ReactDOM' },

    plugins,

    resolve: {
      extensions: ['.js', '.jsx'],
      alias: {
        react: path.resolve(__dirname, 'node_modules/react'),
        'react-dom': path.resolve(__dirname, 'node_modules/react-dom'),
      },
      // optional but helpful if you use `npm link` or a monorepo:
      symlinks: false,
    },

    devtool: isProd ? 'source-map' : 'eval-cheap-module-source-map',

    // Auto rebuild on changes
    watch: true,
    watchOptions: {
      ignored: /node_modules/,
      aggregateTimeout: 300,
    },
  };
};
