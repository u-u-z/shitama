#ifndef CONNECTIONDIALOG_H
#define CONNECTIONDIALOG_H

#include <QDialog>
#include <QMap>
#include <QList>
#include <QChart>
#include <QChartView>

#include <api.h>

namespace Ui {
class ConnectionDialog;
}

class ConnectionDialog : public QDialog
{
    Q_OBJECT

public:
    explicit ConnectionDialog(QWidget *parent = 0);
    ~ConnectionDialog();

private:
    Ui::ConnectionDialog* ui;
    QtCharts::QChart* chart;
    QtCharts::QChartView* chartView;
    int chartIndex;
    QMap<QString, QList<qreal>> chartData;

    API* api;
    int statusWatcher;

    virtual void timerEvent(QTimerEvent* event);

    void updateConnectionStatus();
    void updateDelayChart();
};

#endif // CONNECTIONDIALOG_H
