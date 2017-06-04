#include "mainwindow.h"
#include <QApplication>

int main(int argc, char *argv[])
{

    QCoreApplication::setOrganizationName("Shitama");
    QCoreApplication::setOrganizationDomain("shitama.tldr.run");
    QCoreApplication::setApplicationName("Shitama");

    QApplication a(argc, argv);
    MainWindow w;
    w.show();

    return a.exec();
}
